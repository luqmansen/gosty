FROM nginx:1.19.6-alpine

# Dockerfile for dev, run the build on local
ADD ./web/config/default.conf.template /etc/nginx/templates/default.conf.template
ADD ./web/config/nginx.conf /etc/nginx/nginx.conf
ADD ./web/build /usr/share/nginx/html

HEALTHCHECK --interval=5s --timeout=3s CMD wget -O /dev/null http://localhost || exit 1

EXPOSE 80
