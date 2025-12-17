export default function setupGoogleCompute(widgetElement) {
    const root = widgetElement.querySelector('.gce-widget');
    if (!root) return;

    const widgetId = root.dataset.widgetId;
    const actionContainers = root.querySelectorAll('.gce-actions');

    actionContainers.forEach((container) => {
        container.querySelectorAll('.gce-action-button').forEach((button) => {
            button.addEventListener('click', async () => {
                if (button.disabled) return;

                const action = button.dataset.action;
                const instance = container.dataset.instance;
                const zone = container.dataset.zone;

                if (!action || !instance || !zone) return;

                const previousText = button.textContent;
                widgetElement.style.opacity = '0.6';
                container.querySelectorAll('button').forEach((b) => (b.disabled = true));
                button.textContent = '...';

                try {
                    const response = await fetch(`${pageData.baseURL}/api/google-compute/${widgetId}/action`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            action,
                            instance,
                            zone,
                        }),
                    });

                    if (!response.ok) {
                        throw new Error('Request failed');
                    }

                    const html = await response.text();
                    const tempDiv = document.createElement('div');
                    tempDiv.innerHTML = html;
                    const newWidget = tempDiv.firstElementChild;

                    widgetElement.replaceWith(newWidget);
                    setupGoogleCompute(newWidget);
                } catch (err) {
                    console.error('Failed to perform Compute Engine action', err);
                    widgetElement.style.opacity = '1';
                    container.querySelectorAll('button').forEach((b) => (b.disabled = false));
                    button.textContent = previousText;
                }
            });
        });
    });
}
