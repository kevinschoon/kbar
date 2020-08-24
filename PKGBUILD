pkgname=kbar
pkgver=0.0.1
pkgrel=1
pkgdesc="My personal i3status implementation written with Barista"
url="https://github.com/kevinschoon/kbar"
arch=(x86_64 aarch64 armv7h armv7l)
license=('MIT')
md5sums=()
validpgpkeys=()

build() {
	cp ../bin/kbar .
}

package() {
	install -Dm755 "${pkgname}" -t "${pkgdir}"/usr/bin/
}
