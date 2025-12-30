# Maintainer: Matheus Lasserre <matheuslasserre3@gmail.com>
pkgname=gocrosshair
pkgver=1.0.1
pkgrel=1
pkgdesc="Lightweight, static crosshair overlay for X11/XWayland gaming"
arch=('x86_64')
url="https://github.com/MatheusLasserre/gocrosshair"
license=('MIT')
makedepends=('go' 'git')
source=("$pkgname-$pkgver.tar.gz::${url}/archive/v${pkgver}.tar.gz"
        "gocrosshair.desktop"
        "gocrosshair.png::${url}/raw/main/icon.png")
sha256sums=('958e11b4ee69780041c61bbc9e85b2a7836c4a1a68e2a8e9f2387badeece6ae2'
            'e9a53379e3336b386d1a1cb79e917c6da17d8c6193825343fc551ad1bb84695c'
            'ed24582550439f93f159248e03581bf047ba318cd124f7a11a566d435cd2fa31')

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
