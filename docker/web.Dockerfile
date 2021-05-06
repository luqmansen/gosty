FROM node:15.11.0-alpine3.10 as builder

WORKDIR '/app'

COPY web/package*.json .
RUN npm install

# TODO [#17]: Build react on runner
# Build react on CI runner so we improve built speed
# by using github actions runner caching support
COPY web/ .
RUN npm run build

FROM nginx:1.19.6-alpine

ADD ./web/config/default.conf.template /etc/nginx/templates/default.conf.template
ADD ./web/config/nginx.conf /etc/nginx/nginx.conf

COPY --from=builder /app/build /usr/share/nginx/html

EXPOSE 80
