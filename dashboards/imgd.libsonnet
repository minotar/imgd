

local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local row = grafana.row;
local prometheus = grafana.prometheus;
local template = grafana.template;
local graphPanel = grafana.graphPanel;
local gaugePanel = grafana.gaugePanel;
local heatmapPanel = grafana.heatmapPanel;

{

    imgdCacheInsertHeatmap:
        heatmapPanel.new(
            'Cache Insert Heatmap',
            datasource='$datasource',
            color_mode='opacity',
            dataFormat='tsbuckets',
            yAxis_format='s',
            yAxis_decimals=1,
            maxDataPoints=100,
        )
        .addTarget(prometheus.target(
            'sum(increase(imgd_cache_operation_duration_seconds_bucket{instance=~"$instance", operation="insert"}[$__rate_interval])) by (le)',
            format='heatmap',
            legendFormat='{{ le }}',
        )),

    imgdCacheRetrieveHeatmap:
        heatmapPanel.new(
            'Cache Retrieve Heatmap',
            datasource='$datasource',
            color_mode='opacity',
            dataFormat='tsbuckets',
            yAxis_format='s',
            yAxis_decimals=1,
            maxDataPoints=100,
        )
        .addTarget(prometheus.target(
            'sum(increase(imgd_cache_operation_duration_seconds_bucket{instance=~"$instance", operation="retrieve"}[$__rate_interval])) by (le)',
            format='heatmap',
            legendFormat='{{ le }}',
        )),


    mcclientGetCalls:
        graphPanel.new(
            'McClient API Calls',
            datasource='$datasource',
            //format='ops',
            legend_show=true,
            legend_values=true,
            legend_current=true,
            legend_avg=true,
            legend_alignAsTable=false,
        )
        .addTarget(prometheus.target(
            'sum(irate(imgd_mcclient_api_duration_seconds_count{instance=~"$instance"}[$__rate_interval])) by (source)',
            legendFormat='{{ source }})',
        )),

    mcclientGetErrors:
        graphPanel.new(
            'McClient API Errors',
            datasource='$datasource',
            //format='ops',
            legend_show=true,
            legend_values=true,
            legend_current=true,
            legend_avg=true,
            legend_alignAsTable=false,
        )
        .addTarget(prometheus.target(
            'sum(irate(imgd_mcclient_api_get_errors{instance=~"$instance"}[$__rate_interval])) by (source, event)',
            legendFormat='{{ source }} ({{ event }})',
        )),

    mcclientGetAPIProfileHeatmap:
        heatmapPanel.new(
            'McClient GetAPIProfile Heatmap',
            datasource='$datasource',
            color_mode='opacity',
            dataFormat='tsbuckets',
            yAxis_format='s',
            yAxis_decimals=1,
            maxDataPoints=100,
        )
        .addTarget(prometheus.target(
            'sum(increase(imgd_mcclient_api_duration_seconds_bucket{instance=~"$instance", source="GetAPIProfile"}[$__rate_interval])) by (le)',
            format='heatmap',
            legendFormat='{{ le }}',
        )),

    mcclientGetSessionProfileHeatmap:
        heatmapPanel.new(
            'McClient GetSessionProfile Heatmap',
            datasource='$datasource',
            color_mode='opacity',
            dataFormat='tsbuckets',
            yAxis_format='s',
            yAxis_decimals=1,
            maxDataPoints=100,
        )
        .addTarget(prometheus.target(
            'sum(increase(imgd_mcclient_api_duration_seconds_bucket{instance=~"$instance", source="GetSessionProfile"}[$__rate_interval])) by (le)',
            format='heatmap',
            legendFormat='{{ le }}',
        )),

    // Todo: Maybe use "rate" and a 1m interval
    mcclientCacheStatus:
        graphPanel.new(
            'McClient Cache Hit Rate',
            datasource='$datasource',
            linewidth=2,
            format='ops',
            stack=true,
            legend_show=true,
            legend_values=true,
            legend_avg=true,
            legend_alignAsTable=true,
            legend_rightSide=true,
        )
        .addTarget(prometheus.target(
            'sum(irate(imgd_mcclient_cache_status{instance=~"$instance", status="hit"}[$__rate_interval])) by (cache)',
            legendFormat='{{ cache }} HIT',
        ))
        .addTarget(prometheus.target(
            'sum(irate(imgd_mcclient_cache_status{instance=~"$instance", status="miss"}[$__rate_interval])) by (cache)',
            legendFormat='{{ cache }} MISS',
        ))
        .addSeriesOverride({
                alias: '/.* HIT/',
                stack: 'A',
        })
        .addSeriesOverride({
                alias: '/.* MISS/',
                transform: 'negative-Y',
                stack: 'B',
        }),

    // DRY these gauges??

    mcclientCacheUUIDGauge:
        gaugePanel.new(
            'CacheUUID Hit Rate',
            datasource='$datasource',
            unit='percentunit',
            min=0,
            max=1,
        )
        .addThresholds([
            { color: 'red', value: 0 },
            { color: 'orange', value: 0.35 },
            { color: 'green', value: 0.55 },
        ])
        .addTarget(prometheus.target(
            'sum(rate(imgd_mcclient_cache_status{instance=~"$instance", cache="CacheUUID", status="hit"}[$__rate_interval])) / ignoring(status) sum(rate(imgd_mcclient_cache_status{instance=~"$instance", cache="CacheUUID", status=~"hit|miss"}[$__rate_interval]))'
        )),

    mcclientCacheUserDataGauge:
        gaugePanel.new(
            'CacheUserData Hit Rate',
            datasource='$datasource',
            unit='percentunit',
            min=0,
            max=1,
        )
        .addThresholds([
            { color: 'red', value: 0 },
            { color: 'orange', value: 0.4 },
            { color: 'green', value: 0.6 },
        ])
        .addTarget(prometheus.target(
            'sum(rate(imgd_mcclient_cache_status{instance=~"$instance", cache="CacheUserData", status="hit"}[$__rate_interval])) / ignoring(status) sum(rate(imgd_mcclient_cache_status{instance=~"$instance", cache="CacheUserData", status=~"hit|miss"}[$__rate_interval]))'
        )),

    mcclientCacheTexturesGauge:
        gaugePanel.new(
            'CacheTextures Hit Rate',
            datasource='$datasource',
            unit='percentunit',
            min=0,
            max=1,
        )
        .addThresholds([
            { color: 'red', value: 0 },
            { color: 'orange', value: 0.1 },
            { color: 'green', value: 0.3 },
        ])
        .addTarget(prometheus.target(
            'sum(rate(imgd_mcclient_cache_status{instance=~"$instance", cache="CacheTextures", status="hit"}[$__rate_interval])) / ignoring(status) sum(rate(imgd_mcclient_cache_status{instance=~"$instance", cache="CacheTextures", status=~"hit|miss"}[$__rate_interval]))'
        )),

}

