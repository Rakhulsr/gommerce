FROM node:18-alpine AS frontend_builder


WORKDIR /app

COPY package.json ./
COPY package-lock.json ./


RUN npm ci

COPY tailwind.config.js ./
COPY postcss.config.js ./
COPY src/input.css src/

RUN npm run build


#BE
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /usr/local/bin/toko-bulan-app ./main.go



#image
FROM alpine:latest AS final_app_image

WORKDIR /root/

COPY --from=backend-builder /usr/local/bin/toko-bulan-app .

COPY templates/ ./templates/


COPY --from=frontend_builder /app/static/assets/css/style.css ./static/assets/css/style.css
COPY static/assets/images/ ./static/assets/images/
COPY static/assets/js/ ./static/assets/js/
COPY static/uploads/products/ ./static/uploads/products/



EXPOSE 8080

CMD ["./toko-bulan-app"]