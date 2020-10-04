FROM golang:latest

WORKDIR /app

COPY . .

RUN make --no-print-directory install

CMD make --no-print-directory init
