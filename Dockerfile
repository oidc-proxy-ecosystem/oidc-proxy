FROM golang:1.15-alpine as build
ENV TZ=Asia/Tokyo

WORKDIR /var/app/golang

RUN apk update \
    && apk add make openssh git

COPY . .
RUN go get -v \
    && make bin/proxy/static \
    && make plugin/memory/static

FROM alpine:3
ENV TZ=Asia/Tokyo
ENV LOG_LEVEL=3
WORKDIR /app
COPY --from=build /var/app/golang/bin/proxy /usr/bin/
COPY --from=build /var/app/golang/oidc-plugin/memory /app/oidc-plugin/
COPY docker-entrypoint.sh /usr/bin/
RUN chmod +x /usr/bin/docker-entrypoint.sh \
    && chmod +x /usr/bin/proxy \
    && apk --no-cache add tzdata \
    && cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime \
    && echo "Asia/Tokyo" >  /etc/timezone \
    && rm  -rf /tmp/* /var/cache/apk/*

EXPOSE 8080

# ENTRYPOINT ["docker-entrypoint.sh"]

CMD ["proxy", "proxy", "run" "-c", "application.yaml" ]