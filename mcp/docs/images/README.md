# SlimMCP Dashboard Images

## Grafana Dashboard Screenshot

The `grafana-dashboard.png` file should contain a screenshot of the SlimMCP Grafana dashboard in action.

To generate this image:

1. Import the dashboard from `examples/grafana-dashboard.json` into your Grafana instance
2. Let SlimMCP run for a while to generate metrics
3. Take a screenshot of the dashboard
4. Save it as `grafana-dashboard.png` in this directory

### Expected Dashboard Components

The screenshot should show:
- Token savings overview (percentage and totals)
- Request rate statistics
- Success rate gauge
- Token savings over time graph
- Tokens saved by server breakdown
- Request duration metrics
- Filtering overhead performance
- Request distribution by server and status

### Recommended Image Specifications

- Format: PNG
- Resolution: 1920x1080 or higher
- File size: < 500KB (optimized for README display)
