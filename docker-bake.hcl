variable "GOPROXY" {
  default = "https://goproxy.cn,direct"
}

variable "HTTP_PROXY" {
  default = ""
}

variable "HTTPS_PROXY" {
  default = ""
}

variable "VERSION" {
  default = "unknown"
}

variable "BUILD_TIME" {
  default = "unknown"
}

variable "GIT_COMMIT" {
  default = "unknown"
}

variable "REGISTRY" {
  default = "kcilnk"
}

target "ipaddress-api" {
  context = "."
  dockerfile = "Dockerfile"
  platforms = ["linux/amd64", "linux/arm64"]
  args = {
    VERSION = VERSION
    BUILD_TIME = BUILD_TIME
    GIT_COMMIT = GIT_COMMIT
    GOPROXY = GOPROXY
    HTTP_PROXY = HTTP_PROXY
    HTTPS_PROXY = HTTPS_PROXY
  }
  tags = [
    "${REGISTRY}/go-ipaddress-api:${VERSION}",
    "${REGISTRY}/go-ipaddress-api:latest"
  ]
  cache-from = ["type=gha,scope=go-mod"]
  cache-to = ["type=gha,mode=max,scope=go-mod"]
}

target "ipaddress-api-local" {
  context = "."
  dockerfile = "Dockerfile"
  platforms = ["linux/amd64", "linux/arm64"]
  args = {
    VERSION = VERSION
    BUILD_TIME = BUILD_TIME
    GIT_COMMIT = GIT_COMMIT
    GOPROXY = GOPROXY
    HTTP_PROXY = HTTP_PROXY
    HTTPS_PROXY = HTTPS_PROXY
  }
  tags = [
    "${REGISTRY}/go-ipaddress-api:${VERSION}",
    "${REGISTRY}/go-ipaddress-api:latest"
  ]
  cache-from = ["type=local,src=.buildx-cache"]
  cache-to = ["type=local,mode=max,dest=.buildx-cache-new"]
}
