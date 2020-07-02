FROM registry.svc.ci.openshift.org/openshift/release:golang-1.13 AS builder
WORKDIR /go/src/github.com/eparis/bugtool
COPY . .
ENV GO_PACKAGE github.com/eparis/bugtool
ENV GO_BUILD_PACKAGES ./cmd/blocker-slack
RUN make build --warn-undefined-variables

FROM registry.svc.ci.openshift.org/ocp/4.2:base
COPY --from=builder /go/src/github.com/eparis/bugtool/blocker-slack /usr/bin/
