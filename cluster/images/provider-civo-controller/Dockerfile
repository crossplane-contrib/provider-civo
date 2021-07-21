FROM BASEIMAGE
RUN apk --no-cache add ca-certificates bash

ADD provider /usr/local/bin/crossplane-civo-provider

EXPOSE 8080
USER 1001
ENTRYPOINT ["crossplane-civo-provider"]