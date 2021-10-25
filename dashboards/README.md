
# Grafana Dashboards

To Build the skind dashboard for use in Grafana, use jsonnet.

``` shell
go install github.com/google/go-jsonnet/cmd/jsonnet@latest
go install github.com/google/go-jsonnet/cmd/jsonnetfmt@latest
go install github.com/monitoring-mixins/mixtool/cmd/mixtool@latest
go install github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@latest

# In dashboards/ directory, will load dependancies into dashboards/vendor
jb install

# In main directory
make dashboards/out/skind.json
```
