FROM node:15.11.0-alpine3.10 as builder

WORKDIR '/app'

COPY web/package*.json .
RUN npm install

COPY web/ .
RUN npm run build

FROM nginx:1.19.6-alpine

ADD ./web/config/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/build /usr/share/nginx/html

ADD script/sed.sh /docker-entrypoint.d
RUN apk add jq && chmod +x /docker-entrypoint.d/sed.sh

EXPOSE 80
