module github.com/coreyk/piinky/display-go

go 1.23

toolchain go1.23.4

require (
	github.com/chromedp/chromedp v0.11.2
	periph.io/x/conn/v3 v3.7.1
	periph.io/x/devices/v3 v3.7.1
	periph.io/x/host/v3 v3.8.2
)

replace periph.io/x/devices/v3 => github.com/fstanis/periph-devices/v3 v3.0.0-20240928183903-2a24918d563f

require (
	github.com/chromedp/cdproto v0.0.0-20241022234722-4d5d5faf59fb // indirect
	github.com/chromedp/sysutil v1.1.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	golang.org/x/sys v0.26.0 // indirect
)
