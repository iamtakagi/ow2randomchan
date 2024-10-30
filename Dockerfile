FROM golang:latest AS builder

WORKDIR /work
COPY . /work
RUN go build .

FROM debian:latest AS runner
WORKDIR /bin

COPY --from=builder /work/ow2randomchan /bin/ow2randomchan
COPY --from=builder /work/heroes.json /bin/heroes.json
RUN apt-get update && apt-get install -y \
    ca-certificates \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/bin/ow2randomchan"]