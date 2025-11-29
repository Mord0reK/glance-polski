export default function setupBeszel(widgetElement) {
    console.log('setupBeszel called', widgetElement);
    const chartContainers = widgetElement.querySelectorAll('.beszel-chart-container');
    
    console.log('Found chart containers:', chartContainers.length);
    
    if (chartContainers.length === 0) {
        return;
    }

    chartContainers.forEach(container => {
        initializeChart(container);
    });
}

function initializeChart(container) {
    console.log('initializeChart called', container);
    const widgetId = container.dataset.widgetId;
    const systemId = container.dataset.systemId;
    console.log('widgetId:', widgetId, 'systemId:', systemId);
    
    const metricSelect = container.querySelector('.beszel-metric-select');
    const timeSelect = container.querySelector('.beszel-time-select');
    const chartElement = container.querySelector('.beszel-chart');
    const chartSvg = container.querySelector('.beszel-chart-svg polyline');
    const tooltip = container.querySelector('.beszel-chart-tooltip');
    const cursor = container.querySelector('.beszel-chart-cursor');
    const axisContainer = container.querySelector('.beszel-chart-axis');
    const loadingElement = container.querySelector('.beszel-chart-loading');
    const dataScript = container.querySelector('.beszel-chart-data');
    const yAxisMax = container.querySelector('.beszel-y-max');
    const yAxisMid = container.querySelector('.beszel-y-mid');
    const yAxisMin = container.querySelector('.beszel-y-min');

    if (!widgetId || !systemId) {
        console.error('Beszel chart: brak widgetId lub systemId');
        if (loadingElement) loadingElement.style.display = 'none';
        return;
    }

    let seriesData = [];

    // Funkcja do pobierania danych wykresu
    async function fetchChartData() {
        const metric = metricSelect.value;
        const timeRange = timeSelect.value;
        
        console.log('fetchChartData called', { metric, timeRange, widgetId, systemId });

        loadingElement.style.display = 'flex';
        chartSvg.setAttribute('points', '');

        try {
            const url = `${pageData.baseURL}/api/beszel/${widgetId}/chart`;
            console.log('Fetching from:', url);
            
            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    system_id: systemId,
                    metric: metric,
                    time_range: timeRange
                }),
            });

            console.log('Response status:', response.status);
            
            if (!response.ok) {
                const errorText = await response.text();
                console.error('Nie udało się pobrać danych wykresu:', errorText);
                loadingElement.style.display = 'none';
                return;
            }

            const data = await response.json();
            console.log('Received data:', data);
            seriesData = data.series || [];
            
            // Aktualizacja wykresu
            if (data.points) {
                console.log('Setting points:', data.points.substring(0, 100) + '...');
                chartSvg.setAttribute('points', data.points);
            }

            // Aktualizacja etykiet osi Y z min/max z backendu
            const minValue = data.minValue !== undefined ? data.minValue : 0;
            const maxValue = data.maxValue !== undefined ? data.maxValue : 100;
            updateYAxisLabels(metric, minValue, maxValue);

            // Aktualizacja etykiet osi X
            updateAxisLabels(axisContainer, data.axisLabels || []);

            // Zapisz dane do scriptu dla tooltip
            dataScript.textContent = JSON.stringify(seriesData);

        } catch (error) {
            console.error('Błąd podczas pobierania danych wykresu:', error);
        } finally {
            loadingElement.style.display = 'none';
        }
    }

    // Funkcja do aktualizacji etykiet osi Y
    function updateYAxisLabels(metric, minValue, maxValue) {
        if (!yAxisMax || !yAxisMid || !yAxisMin) return;
        
        const midValue = (minValue + maxValue) / 2;
        
        if (metric === 'network') {
            // Dla sieci - wartości w MB
            yAxisMax.textContent = maxValue.toFixed(1) + ' MB';
            yAxisMid.textContent = midValue.toFixed(1) + ' MB';
            yAxisMin.textContent = minValue.toFixed(1) + ' MB';
        } else {
            // Dla CPU, RAM, Disk - procentowe wartości
            yAxisMax.textContent = maxValue.toFixed(0) + '%';
            yAxisMid.textContent = midValue.toFixed(0) + '%';
            yAxisMin.textContent = minValue.toFixed(0) + '%';
        }
    }

    // Funkcja do aktualizacji etykiet osi X
    function updateAxisLabels(axisContainer, labels) {
        axisContainer.innerHTML = '';
        labels.forEach(label => {
            const labelEl = document.createElement('div');
            labelEl.className = 'text-compact color-subdue beszel-axis-label';
            labelEl.style.position = 'absolute';
            labelEl.style.left = `${label.left}%`;
            labelEl.style.transform = label.transform;
            labelEl.style.whiteSpace = 'nowrap';
            labelEl.style.fontSize = 'var(--font-size-h6)';
            labelEl.textContent = label.label;
            axisContainer.appendChild(labelEl);
        });
    }

    // Event listenery dla selectów
    metricSelect.addEventListener('change', fetchChartData);
    timeSelect.addEventListener('change', fetchChartData);

    // Interakcja z wykresem (tooltip i cursor)
    chartElement.addEventListener('mousemove', (e) => {
        if (seriesData.length === 0) return;

        const rect = chartElement.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const width = rect.width;

        let index = Math.round((x / width) * (seriesData.length - 1));
        index = Math.max(0, Math.min(index, seriesData.length - 1));

        const point = seriesData[index];
        if (!point) return;

        const metric = metricSelect.value;
        let unit = '%';
        let valueDisplay = point.value.toFixed(1);

        if (metric === 'network') {
            unit = ' MB';
            valueDisplay = point.value.toFixed(2);
        }

        // Aktualizacja tooltip
        tooltip.innerHTML = `
            <div class="text-label" style="margin-bottom: 0.2rem;">${point.label}</div>
            <div class="color-highlight">${valueDisplay}${unit}</div>
        `;

        // Pozycjonowanie tooltip
        const tooltipRect = tooltip.getBoundingClientRect();
        let tooltipLeft = x - (tooltipRect.width / 2);

        if (tooltipLeft < 0) tooltipLeft = 0;
        if (tooltipLeft + tooltipRect.width > width) tooltipLeft = width - tooltipRect.width;

        tooltip.style.left = `${tooltipLeft}px`;
        tooltip.style.top = `-${tooltipRect.height + 5}px`;
        tooltip.style.opacity = '1';

        // Pozycjonowanie cursor
        const pointX = (index / (seriesData.length - 1)) * width;
        cursor.style.left = `${pointX}px`;
        cursor.style.opacity = '1';
    });

    chartElement.addEventListener('mouseleave', () => {
        tooltip.style.opacity = '0';
        cursor.style.opacity = '0';
    });

    // Ładowanie początkowych danych
    fetchChartData();
}
