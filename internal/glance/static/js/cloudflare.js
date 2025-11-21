export default function setupCloudflare(widgetElement) {
    const contentDiv = widgetElement.querySelector('[data-widget-id]');
    if (!contentDiv) return;

    const widgetId = contentDiv.dataset.widgetId;
    const select = widgetElement.querySelector('select.cloudflare-time-range');
    
    // Setup select change handler
    if (select) {
        select.addEventListener('change', async (e) => {
            const timeRange = e.target.value;
            
            widgetElement.style.opacity = '0.5';
            select.disabled = true;

            try {
                const response = await fetch(`${pageData.baseURL}/api/cloudflare/${widgetId}/update`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ time_range: timeRange }),
                });

                if (!response.ok) {
                    console.error('Failed to update cloudflare widget');
                    widgetElement.style.opacity = '1';
                    select.disabled = false;
                    return;
                }

                const html = await response.text();
                
                const tempDiv = document.createElement('div');
                tempDiv.innerHTML = html;
                const newWidget = tempDiv.firstElementChild;
                
                widgetElement.replaceWith(newWidget);
                
                // Re-initialize the new widget
                setupCloudflare(newWidget);
                
            } catch (error) {
                console.error('Error updating cloudflare widget:', error);
                widgetElement.style.opacity = '1';
                select.disabled = false;
            }
        });
    }

    // Setup chart interaction
    const chartContainer = widgetElement.querySelector('.chart-container');
    const dataScript = widgetElement.querySelector('.cloudflare-data');
    const tooltip = widgetElement.querySelector('.chart-tooltip');
    const cursor = widgetElement.querySelector('.chart-cursor');

    if (chartContainer && dataScript && tooltip && cursor) {
        let seriesData = [];
        try {
            seriesData = JSON.parse(dataScript.textContent);
        } catch (e) {
            console.error('Failed to parse cloudflare data', e);
            return;
        }

        if (!seriesData || seriesData.length === 0) return;

        chartContainer.addEventListener('mousemove', (e) => {
            const rect = chartContainer.getBoundingClientRect();
            const x = e.clientX - rect.left;
            const width = rect.width;
            
            // Calculate index based on x position
            // The points are distributed evenly across the width
            // index = round((x / width) * (length - 1))
            
            let index = Math.round((x / width) * (seriesData.length - 1));
            index = Math.max(0, Math.min(index, seriesData.length - 1));
            
            const point = seriesData[index];
            
            // Update tooltip content
            tooltip.innerHTML = `
                <div class="text-label">${point.label}</div>
                <div class="color-highlight">${point.requests} Zapyt.</div>
                <div class="color-subdue text-very-compact">${point.uniques} Unik.</div>
            `;
            
            // Position tooltip
            // Try to center it above the cursor, but keep it within bounds
            const tooltipRect = tooltip.getBoundingClientRect();
            
            let tooltipLeft = x - (tooltipRect.width / 2);
            
            // Clamp to container bounds
            if (tooltipLeft < 0) tooltipLeft = 0;
            if (tooltipLeft + tooltipRect.width > width) tooltipLeft = width - tooltipRect.width;
            
            tooltip.style.left = `${tooltipLeft}px`;
            tooltip.style.top = `-${tooltipRect.height + 5}px`; // Position above
            tooltip.style.opacity = '1';
            
            // Position cursor
            const pointX = (index / (seriesData.length - 1)) * width;
            cursor.style.left = `${pointX}px`;
            cursor.style.opacity = '1';
        });

        chartContainer.addEventListener('mouseleave', () => {
            tooltip.style.opacity = '0';
            cursor.style.opacity = '0';
        });
    }
}
