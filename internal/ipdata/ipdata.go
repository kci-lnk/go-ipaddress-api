package ipdata

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

type IPData struct {
	v4 *xdb.Searcher
	v6 *xdb.Searcher
}

var (
	instance *IPData
	once     sync.Once
)

func New(dbPath string) (*IPData, error) {
	v4Path := filepath.Join(dbPath, "base_full_v4.xdb")
	v6Path := filepath.Join(dbPath, "base_full_v6.xdb")

	v4, err := xdb.NewWithFileOnly(xdb.IPv4, v4Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open IPv4 database: %w", err)
	}

	v6, err := xdb.NewWithFileOnly(xdb.IPv6, v6Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open IPv6 database: %w", err)
	}

	return &IPData{v4: v4, v6: v6}, nil
}

func NewWithCache(dbPath string) (*IPData, error) {
	v4Path := filepath.Join(dbPath, "base_full_v4.xdb")
	v6Path := filepath.Join(dbPath, "base_full_v6.xdb")

	v4Index, err := xdb.LoadVectorIndexFromFile(v4Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load IPv4 vector index: %w", err)
	}

	v4, err := xdb.NewWithVectorIndex(xdb.IPv4, v4Path, v4Index)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPv4 searcher: %w", err)
	}

	v6Index, err := xdb.LoadVectorIndexFromFile(v6Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load IPv6 vector index: %w", err)
	}

	v6, err := xdb.NewWithVectorIndex(xdb.IPv6, v6Path, v6Index)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPv6 searcher: %w", err)
	}

	return &IPData{v4: v4, v6: v6}, nil
}

func Get() *IPData {
	return instance
}

func Set(d *IPData) {
	instance = d
}

type LookupResult struct {
	Version     string
	Continent   string
	Country     string
	Province    string
	City        string
	District    string
	Isp         string
	CountryCode string
	Fields      []string
	Raw         string
}

func (d *IPData) Lookup(ip string) (*LookupResult, error) {
	isV6 := isIPv6(ip)

	var region string
	var err error

	if isV6 && d.v6 != nil {
		region, err = d.v6.Search(ip)
	} else if d.v4 != nil {
		region, err = d.v4.Search(ip)
	}

	if err != nil {
		return nil, fmt.Errorf("lookup failed: %w", err)
	}

	if region == "" {
		return nil, nil
	}

	return parseRegion(region, isV6), nil
}

func parseRegion(region string, isV6 bool) *LookupResult {
	parts := strings.Split(region, "|")

	get := func(i int) string {
		if i >= len(parts) {
			return ""
		}
		v := parts[i]
		if v == "0" || v == "" {
			return ""
		}
		return v
	}

	// ip2region xdb raw data format (this specific xdb file):
	// 0:continent|1:country|2:province|3:city|4:district|5:isp|6:lat|7:lon|8:zipcode|9:areacode|10:iddcode|11:timezone|12:currency|13:country_code|14:mcc|15:mnc|16:mobile|17:icncode|18:datacenter|19:country_extended
	continent := get(0)
	country := get(1)
	province := get(2)
	city := get(3)
	district := get(4)
	isp := get(5)
	lat := get(6)
	lon := get(7)
	zipcode := get(8)
	areacode := get(9)
	iddcode := get(10)
	timezone := get(11)
	currency := get(12)
	countryCode := get(13)

	fields := []string{
		continent,   // 0: continent
		country,     // 1: country
		province,    // 2: province
		city,        // 3: city
		district,    // 4: district
		isp,         // 5: isp
		lat,         // 6: latitude
		lon,         // 7: longitude
		zipcode,     // 8: zipcode
		areacode,    // 9: areacode
		iddcode,     // 10: iddcode
		timezone,    // 11: timezone
		countryCode, // 12: country_code
		currency,    // 13: currency
		"",          // 14: weather code
		"",          // 15: weather name
		"",          // 16: mcc
		"",          // 17: mnc
		"",          // 18: mobile brand
		"",          // 19: icncode
		"",          // 20: datacenter
		"",          // 21: country extended
	}

	return &LookupResult{
		Version:     boolToVersion(isV6),
		Continent:   continent,
		Country:     country,
		Province:    province,
		City:        city,
		District:    district,
		Isp:         isp,
		CountryCode: countryCode,
		Fields:      fields,
		Raw:         region,
	}
}

func boolToVersion(isV6 bool) string {
	if isV6 {
		return "ipv6"
	}
	return "ipv4"
}

func isIPv6(ip string) bool {
	for i := 0; i < len(ip); i++ {
		if ip[i] == ':' {
			return true
		}
	}
	return false
}
