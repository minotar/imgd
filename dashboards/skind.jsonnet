local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local row = grafana.row;
local prometheus = grafana.prometheus;
local template = grafana.template;
local graphPanel = grafana.graphPanel;
local tablePanel = grafana.tablePanel;
local statPanel = grafana.statPanel;

local _config = (import 'config.libsonnet')._config;

local skindPanel = import 'skind.libsonnet';
local imgdPanel = import 'imgd.libsonnet';
local goDash = import 'go-runtime-mixin/dashboards/go-runtime.json';


{

    'go-runtime.json': goDash,

    'skind.json':

      dashboard.new(
        '%(dashboardNamePrefix)sSkind' % _config,
        editable=true,
        time_from='now-1h',
        uid=(_config.grafanaDashboardIDs['skind.json']),
        tags=(_config.dashboardTags),
      ).addTemplate(
        {
          current: {
            text: 'default',
            value: _config.datasourceName,
          },
          hide: 0,
          label: null,
          name: 'datasource',
          options: [],
          query: 'prometheus',
          refresh: 1,
          regex: _config.datasourceFilterRegex,
          type: 'datasource',
        },
      )
      .addTemplate(
        template.new(
          'instance',
          '$datasource',
          'label_values(up{%(skindSelector)s}, instance)' % _config,
          label='Instance',
          refresh='time',
          includeAll=true,
          sort=1,
        )
      )
      .addPanel(skindPanel.skindInflightRequests,            gridPos={ h: 8, w: 12, x:  0, y: 0 })
      .addPanel(imgdPanel.mcclientCacheUUIDGauge,            gridPos={ h: 8, w: 4,  x: 12, y: 0 })
      .addPanel(imgdPanel.mcclientCacheUserDataGauge,        gridPos={ h: 8, w: 4,  x: 16, y: 0 })
      .addPanel(imgdPanel.mcclientCacheTexturesGauge,        gridPos={ h: 8, w: 4,  x: 20, y: 0 })

      .addPanel(skindPanel.skindHttpStatusRequests,          gridPos={ h: 8, w: 12, x:  0, y: 8 })
      .addPanel(skindPanel.skindHttpRouteHeatmap,            gridPos={ h: 8, w: 12, x: 12, y: 8 })

      .addPanel(skindPanel.skindHttpRouteRequests,           gridPos={ h: 8, w: 12, x:  0, y: 16 })
      .addPanel(imgdPanel.mcclientCacheStatus,               gridPos={ h: 8, w: 12, x: 12, y: 16 })

      .addPanel(imgdPanel.mcclientGetCalls,                  gridPos={ h: 8, w: 12, x:  0, y: 24 })
      .addPanel(imgdPanel.mcclientGetErrors,                 gridPos={ h: 8, w: 12, x: 12, y: 24 })

      .addPanel(imgdPanel.mcclientGetAPIProfileHeatmap,      gridPos={ h: 8, w: 12, x:  0, y: 32 })
      .addPanel(imgdPanel.mcclientGetSessionProfileHeatmap,  gridPos={ h: 8, w: 12, x: 12, y: 32 })

      .addPanel(imgdPanel.imgdCacheInsertHeatmap,            gridPos={ h: 8, w: 12, x:  0, y: 40 })
      .addPanel(imgdPanel.imgdCacheRetrieveHeatmap,          gridPos={ h: 8, w: 12, x: 12, y: 40 })

      + { refresh: _config.refresh },
}
