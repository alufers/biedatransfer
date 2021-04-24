FROM archlinux:base-devel
RUN pacman -Sy --noconfirm && pacman -S --noconfirm binwalk file go
WORKDIR /go/src/app
COPY . .

RUN  go build .

CMD ["/go/src/app/biedatransfer"]
