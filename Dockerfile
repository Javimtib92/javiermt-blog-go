ARG GOLANG_VERSION=1.22.0

FROM golang:${GOLANG_VERSION}-bookworm AS build-stage

ENV GOPROXY=https://proxy.golang.org

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

ARG VIPS_VERSION=8.15.1

ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig

# Installs libvips + required libraries
# Installs libvips + required libraries
RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update && \
  apt-get install --no-install-recommends -y \
  ca-certificates \
  automake build-essential curl \
  meson ninja-build pkg-config \
  gobject-introspection gtk-doc-tools libglib2.0-dev libjpeg62-turbo-dev libpng-dev \
  libwebp-dev libtiff-dev libexif-dev libxml2-dev libpoppler-glib-dev \
  swig libpango1.0-dev libmatio-dev libopenslide-dev libcfitsio-dev libopenjp2-7-dev liblcms2-dev \
  libgsf-1-dev libfftw3-dev liborc-0.4-dev librsvg2-dev libimagequant-dev libaom-dev \
  libheif-dev libspng-dev libcgif-dev && \
  cd /tmp && \
    curl -fsSLO https://github.com/libvips/libvips/releases/download/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.xz && \
    tar xf vips-${VIPS_VERSION}.tar.xz && \
    cd vips-${VIPS_VERSION} && \
    meson setup _build \
    --buildtype=release \
    --strip \
    --prefix=/usr/local \
    --libdir=lib \
    -Dgtk_doc=false \
    -Dmagick=disabled \
    -Dintrospection=disabled && \
    ninja -C _build && \
    ninja -C _build install && \
  ldconfig && \
  rm -rf /usr/local/lib/libvips-cpp.* && \
  rm -rf /usr/local/lib/*.a && \
  rm -rf /usr/local/lib/*.la

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/gomod-cache go mod download

COPY . ./

RUN go build -o /javiermt-blog

FROM debian:bookworm-slim

WORKDIR /

COPY --from=build-stage /usr/local/lib /usr/local/lib
COPY --from=build-stage /etc/ssl/certs /etc/ssl/certs

# Install runtime dependencies
RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update && \
  apt-get install --no-install-recommends -y \
  procps libglib2.0-0 libjpeg62-turbo libpng16-16 libopenexr-3-1-30 \
  libwebp7 libwebpmux3 libwebpdemux2 libtiff6 libexif12 libxml2 libpoppler-glib8 \
  libpango1.0-0 libmatio11 libopenslide0 libopenjp2-7 libjemalloc2 \
  libgsf-1-114 libfftw3-bin liborc-0.4-0 librsvg2-2 libcfitsio10 libimagequant0 libaom3 libheif1 \
  libspng0 libcgif0 && \
  ln -s /usr/lib/$(uname -m)-linux-gnu/libjemalloc.so.2 /usr/local/lib/libjemalloc.so && \
  apt-get autoremove -y && \
  apt-get autoclean && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=build-stage /javiermt-blog /javiermt-blog

ENV VIPS_WARNING=0
ENV MALLOC_ARENA_MAX=2
ENV LD_PRELOAD=/usr/local/lib/libjemalloc.so

EXPOSE 8080

USER nobody

ENTRYPOINT ["/javiermt-blog"]