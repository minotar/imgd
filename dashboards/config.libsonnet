{
  _config+:: {

    skindSelector: 'job="skind"',

    dashboardNamePrefix: 'Minotar / ',
    dashboardTags: ['minotar-imgd'],

    grafanaDashboardIDs: {
      // echo -n "skind.json" | md5sum
      'skind.json': '5c3b0d6edfd389ef23d377d5e7ab0f64',
    },

    // The default refresh time for all dashboards, default to 10s
    refresh: '10s',

    datasourceName: 'default',
    // Datasource instance filter regex
    datasourceFilterRegex: '',
  },
}
