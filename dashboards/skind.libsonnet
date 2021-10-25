

local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local row = grafana.row;
local prometheus = grafana.prometheus;
local template = grafana.template;
local graphPanel = grafana.graphPanel;
local gaugePanel = grafana.gaugePanel;
local heatmapPanel = grafana.heatmapPanel;

{
    routeIgnore:: 'route!="metrics", route!="other"',

    skindInflightRequests:
        graphPanel.new(
            'Inflight Requests',
            datasource='$datasource',
            min=0,
            legend_show=true,
            legend_alignAsTable=false,
        )
        .addTarget(prometheus.target(
            'sum(skind_inflight_requests{instance=~"$instance", %(routeIgnore)s}) by (route)' % self,
            legendFormat='{{ route }}',
        ))
        .addTarget(prometheus.target(
            'sum(imgd_mcclient_api_inflight_requests{instance=~"$instance"}) by (source)',
            legendFormat='McClient {{ source }}',
        ))
        .addSeriesOverride({
                alias: '/McClient.*/',
                dashes: true,
                fill: 0,
                yaxis: 2,
        }),


    skindHttpStatusRequests:
        graphPanel.new(
            'HTTP Requests by Status Code',
            datasource='$datasource',
            //format='ops',
            legend_show=true,
            legend_values=true,
            legend_current=true,
            legend_avg=true,
            legend_alignAsTable=false,
        )
        .addTarget(prometheus.target(
            'sum(irate(skind_request_duration_seconds_count{instance=~"$instance", %(routeIgnore)s}[$__rate_interval])) by (status_code)' % self,
            legendFormat='{{ status_code }}',
        )),

    skindHttpRouteRequests:
        graphPanel.new(
            'HTTP Requests by Route',
            datasource='$datasource',
            //format='ops',
            legend_show=true,
            legend_hideZero=true,
            legend_values=true,
            legend_current=true,
            legend_avg=true,
            legend_alignAsTable=false,
        )
        .addTarget(prometheus.target(
            'sum(irate(skind_request_duration_seconds_count{instance=~"$instance", %(routeIgnore)s}[$__rate_interval])) by (route)' % self,
            legendFormat='{{ route }}',
        )),

    skindHttpRouteHeatmap:
        heatmapPanel.new(
            'HTTP Requests Heatmap',
            datasource='$datasource',
            color_mode='opacity',
            dataFormat='tsbuckets',
            yAxis_format='s',
            yAxis_decimals=1,
            maxDataPoints=100,
        )
        .addTarget(prometheus.target(
            'sum(increase(skind_request_duration_seconds_bucket{instance=~"$instance", %(routeIgnore)s}[$__rate_interval])) by (le)' % self,
            format='heatmap',
            legendFormat='{{ le }}',
        )),


}

