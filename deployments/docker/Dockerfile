FROM golang:1.12 AS testservice-dependencies
ENV NAME "testservice"
WORKDIR /opt/${NAME}
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM testservice-dependencies AS build
ENV NAME "testservice"
WORKDIR /opt/${NAME}
COPY . .
RUN ["make", "build"]

FROM alpine
ARG ENV
ARG VERSION
ARG ELASTIC_APM_SERVER_URL
ARG ELASTIC_APM_SECRET_TOKEN
ENV NAME "testservice"
ENV ELASTIC_APM_ENVIRONMENT ${ENV}
ENV ELASTIC_APM_SERVICE_NAME ${NAME}
ENV ELASTIC_APM_SERVICE_VERSION ${VERSION}
ENV ELASTIC_APM_SERVER_URL ${ELASTIC_APM_SERVER_URL}
ENV ELASTIC_APM_SECRET_TOKEN ${ELASTIC_APM_SECRET_TOKEN}
WORKDIR /opt/${NAME}
COPY --from=build /opt/${NAME}/configs/*.yml ./configs/
COPY --from=build /opt/${NAME}/bin/${NAME} ./${NAME}
CMD ./${NAME}
