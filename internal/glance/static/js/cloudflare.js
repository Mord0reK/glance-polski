export default function setupCloudflare(widgetElement) {
    const contentDiv = widgetElement.querySelector('[data-widget-id]');
    if (!contentDiv) {
        console.error('Cloudflare: contentDiv not found');
        return;
    }

    // Setup chart interaction
    const chartContainer = widgetElement.querySelector('.chart-container');
    const dataScript = widgetElement.querySelector('.cloudflare-data');
    const tooltip = widgetElement.querySelector('.chart-tooltip');
    const cursor = widgetElement.querySelector('.chart-cursor');

    console.log('Cloudflare setup:', { chartContainer, dataScript, tooltip, cursor });

    if (!chartContainer || !dataScript || !tooltip || !cursor) {
        console.error('Cloudflare: missing elements', { chartContainer: !!chartContainer, dataScript: !!dataScript, tooltip: !!tooltip, cursor: !!cursor });
        return;
    }

    let seriesData = [];
    try {
        seriesData = JSON.parse(dataScript.textContent);
        console.log('Cloudflare seriesData:', seriesData.length, 'points');
    } catch (e) {
        console.error('Failed to parse cloudflare data', e);
        return;
    }

    if (!seriesData || seriesData.length === 0) {
        console.error('Cloudflare: no series data');
        return;
    }

    chartContainer.addEventListener('mousemove', (e) => {
        const rect = chartContainer.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const width = rect.width;
        
        let index = Math.round((x / width) * (seriesData.length - 1));
        index = Math.max(0, Math.min(index, seriesData.length - 1));
        
        const point = seriesData[index];
        
        const requests = point.requests !== undefined ? point.requests : point.Requests;
        const cachedRequests = point.cachedRequests !== undefined ? point.cachedRequests : point.CachedRequests;
        const threats = point.threats !== undefined ? point.threats : point.Threats;
        const timestamp = point.timestamp !== undefined ? point.timestamp : point.Timestamp;
        const label = point.label !== undefined ? point.label : point.Label;
        
        let timeLabel = label;
        if (timestamp) {
            const ts = timestamp;
            if (ts.length >= 19) {
                try {
                    const date = new Date(ts);
                    const hours = date.getHours().toString().padStart(2, '0');
                    const minutes = date.getMinutes().toString().padStart(2, '0');
                    timeLabel = hours + ':' + minutes;
                } catch (e) {
                    timeLabel = ts.substring(11, 16);
                }
            }
        }
        
        tooltip.innerHTML = `
            <div class="text-label">${timeLabel}</div>
            <div class="color-highlight">${requests} Zapyt.</div>
            <div class="color-subdue text-very-compact">${cachedRequests} Z cache</div>
            <div class="color-subdue text-very-compact">${threats} Zablokowane</div>
        `;
        
        // Make tooltip visible temporarily to get dimensions
        tooltip.style.visibility = 'hidden';
        tooltip.style.opacity = '1';
        
        const tooltipRect = tooltip.getBoundingClientRect();
        
        let tooltipLeft = x - (tooltipRect.width / 2);
        
        if (tooltipLeft < 0) tooltipLeft = 0;
        if (tooltipLeft + tooltipRect.width > width) tooltipLeft = width - tooltipRect.width;
        
        tooltip.style.left = `${tooltipLeft}px`;
        tooltip.style.top = `-${tooltipRect.height + 5}px`;
        
        // Now hide it again - we handle opacity separately
        tooltip.style.visibility = 'visible';
        
        const pointX = (index / (seriesData.length - 1)) * width;
        cursor.style.left = `${pointX}px`;
        cursor.style.opacity = '1';
    });

    chartContainer.addEventListener('mouseleave', () => {
        tooltip.style.opacity = '0';
        cursor.style.opacity = '0';
    });
}
