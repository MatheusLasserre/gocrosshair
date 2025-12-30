# Maintainer: Matheus Lasserre <matheuslasserre3@gmail.com>
pkgname=gocrosshair
pkgver=1.0.0
pkgrel=1
pkgdesc="Lightweight, static crosshair overlay for X11/XWayland gaming"
arch=('x86_64')
url="https://github.com/MatheusLasserre/gocrosshair"
license=('MIT')
makedepends=('go' 'git')
source=("$pkgname-$pkgver.tar.gz::${url}/archive/v${pkgver}.tar.gz"
        "gocrosshair.desktop"
        "gocrosshair.png::${url}/raw/main/icon.png")
# Run 'updpkgsums' to generate real checksums before AUR submission
sha256sums=('SKIP'
            'SKIP'
            'SKIP')

prepare() {
  cd "$pkgname-$pkgver"
  mkdir -p build
}

build() {
  cd "$pkgname-$pkgver"
  export CGO_ENABLED=0
  export GOFLAGS="-buildmode=pie -trimpath -mod=readonly -modcacherw"
  go build -ldflags "-s -w -X main.version=${pkgver}" -o build/gocrosshair .
}

check() {
  cd "$pkgname-$pkgver"
  go test ./...
}

package() {
  cd "$pkgname-$pkgver"
  install -Dm755 build/gocrosshair "$pkgdir/usr/bin/gocrosshair"
  
  if [[ -f LICENSE ]]; then
    install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
  fi
  
  cd ..
  install -Dm644 gocrosshair.desktop "$pkgdir/usr/share/applications/gocrosshair.desktop"
  install -Dm644 gocrosshair.png "$pkgdir/usr/share/pixmaps/gocrosshair.png"
}
