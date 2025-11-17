// Vikunja widget interactivity
function initVikunjaWidget(widget) {
    if (!widget) return;
    
    {
        const widgetID = widget.querySelector('.vikunja-table')?.dataset.widgetId || 
                         widget.querySelector('.vikunja-add-btn')?.dataset.widgetId;
        if (!widgetID) return;

        // Handle add button (can be in table header or empty state)
        const addBtn = widget.querySelector('.vikunja-add-btn');
        if (addBtn) {
            addBtn.addEventListener('click', function() {
                const btnWidgetID = this.dataset.widgetId || widgetID;
                openCreateModal(btnWidgetID);
            });
        }

        // Handle task completion checkboxes
        widget.querySelectorAll('.vikunja-task-checkbox').forEach(checkbox => {
            checkbox.addEventListener('change', function(e) {
                if (!this.checked) {
                    this.checked = false;
                    return;
                }

                const row = this.closest('tr');
                const taskID = parseInt(row.dataset.taskId);

                // if (!confirm('Czy na pewno chcesz oznaczyć to zadanie jako wykonane?')) {
                //    this.checked = false;
                //    return;
                //}

                // Call API to complete task
                fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/complete-task`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ task_id: taskID })
                })
                .then(response => {
                    if (!response.ok) throw new Error('Failed to complete task');
                    return response.json();
                })
                .then(data => {
                    // Play completion sound
                    const completionSound = document.getElementById('vikunja-completion-sound');
                    if (completionSound) {
                        completionSound.currentTime = 0;
                        completionSound.play().catch(err => {
                            console.log('Could not play completion sound:', err);
                        });
                    }
                    
                    // Remove the row with animation
                    row.style.transition = 'opacity 0.3s ease';
                    row.style.opacity = '0';
                    setTimeout(async () => {
                        // Refresh the widget
                        try {
                            const refreshResponse = await fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/refresh`);
                            if (refreshResponse.ok) {
                                const newHTML = await refreshResponse.text();
                                const widgetContainer = document.querySelector(`.widget-type-vikunja`);
                                if (widgetContainer) {
                                    const temp = document.createElement('div');
                                    temp.innerHTML = newHTML;
                                    const newWidget = temp.firstElementChild;
                                    widgetContainer.replaceWith(newWidget);
                                    initVikunjaWidget(newWidget);
                                }
                            }
                        } catch (error) {
                            console.error('Error refreshing widget:', error);
                            // Fallback to just removing the row
                            row.remove();
                        }
                    }, 300);
                })
                .catch(error => {
                    console.error('Error completing task:', error);
                    alert('Nie udało się oznaczyć zadania jako wykonane');
                    this.checked = false;
                });
            });
        });

        // Handle edit buttons
        widget.querySelectorAll('.vikunja-edit-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                const row = this.closest('tr');
                const taskID = parseInt(row.dataset.taskId);
                const taskTitle = this.dataset.taskTitle;
                const taskDueDate = this.dataset.taskDueDate;
                
                // Get current labels
                const currentLabels = Array.from(row.querySelectorAll('.label')).map(label => 
                    parseInt(label.dataset.labelId)
                );

                openEditModal(widgetID, taskID, taskTitle, taskDueDate, currentLabels, row);
            });
        });
    }
}

function openEditModal(widgetID, taskID, title, dueDate, currentLabelIDs, row) {
    const modal = document.getElementById('vikunja-edit-modal');
    const titleInput = document.getElementById('vikunja-edit-title');
    const dueDateInput = document.getElementById('vikunja-edit-due-date');
    const labelsContainer = document.getElementById('vikunja-labels-container');

    // Set current values
    titleInput.value = title || '';
    
    // Convert date format from "2006-01-02 15:04" to datetime-local format "2006-01-02T15:04"
    if (dueDate) {
        dueDateInput.value = dueDate.replace(' ', 'T');
    } else {
        dueDateInput.value = '';
    }

    // Fetch and display labels
    labelsContainer.innerHTML = '<p>Ładowanie etykiet...</p>';
    
    fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/labels`)
        .then(response => response.json())
        .then(labels => {
            labelsContainer.innerHTML = '';
            
            if (labels && labels.length > 0) {
                labels.forEach(label => {
                    const labelCheckbox = document.createElement('label');
                    labelCheckbox.className = 'vikunja-label-option';
                    
                    const color = label.hex_color && label.hex_color[0] !== '#' 
                        ? '#' + label.hex_color 
                        : label.hex_color || '#666';
                    
                    const isChecked = currentLabelIDs.includes(label.id);
                    
                    labelCheckbox.innerHTML = `
                        <input type="checkbox" value="${label.id}" ${isChecked ? 'checked' : ''}>
                        <span class="label" style="border-color: ${color}; color: ${color};">${label.title}</span>
                    `;
                    
                    labelsContainer.appendChild(labelCheckbox);
                });
            } else {
                labelsContainer.innerHTML = '<p>Brak dostępnych etykiet</p>';
            }
        })
        .catch(error => {
            console.error('Error fetching labels:', error);
            labelsContainer.innerHTML = '<p>Nie udało się załadować etykiet</p>';
        });

    modal.style.display = 'flex';

    // Handle close button
    const closeBtn = modal.querySelector('.vikunja-modal-close');
    const cancelBtn = modal.querySelector('.vikunja-btn-cancel');
    const saveBtn = modal.querySelector('.vikunja-btn-save');

    function closeModal() {
        modal.style.display = 'none';
        // Remove event listeners
        closeBtn.removeEventListener('click', closeModal);
        cancelBtn.removeEventListener('click', closeModal);
        saveBtn.removeEventListener('click', saveTask);
    }

    async function saveTask() {
        const newTitle = titleInput.value.trim();
        const newDueDate = dueDateInput.value;
        
        // Get selected label IDs
        const selectedLabels = Array.from(labelsContainer.querySelectorAll('input[type="checkbox"]:checked'))
            .map(checkbox => parseInt(checkbox.value));

        if (!newTitle) {
            alert('Tytuł zadania nie może być pusty');
            return;
        }

        // Convert datetime-local format to RFC3339
        let formattedDueDate = '';
        if (newDueDate) {
            const date = new Date(newDueDate);
            formattedDueDate = date.toISOString();
        }

        try {
            // Step 1: Update title and due date
            const updateResponse = await fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/update-task`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    task_id: taskID,
                    title: newTitle,
                    due_date: formattedDueDate
                })
            });

            if (!updateResponse.ok) {
                throw new Error('Failed to update task');
            }

            // Step 2: Update labels
            // Get current label IDs from the row
            const currentLabels = Array.from(row.querySelectorAll('.label')).map(label => 
                parseInt(label.dataset.labelId)
            );

            // Determine which labels to add and remove
            const labelsToAdd = selectedLabels.filter(id => !currentLabels.includes(id));
            const labelsToRemove = currentLabels.filter(id => !selectedLabels.includes(id));

            // Add new labels
            for (const labelID of labelsToAdd) {
                const addResponse = await fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/add-label`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        task_id: taskID,
                        label_id: labelID
                    })
                });

                if (!addResponse.ok) {
                    throw new Error(`Failed to add label ${labelID}`);
                }
            }

            // Remove old labels
            for (const labelID of labelsToRemove) {
                const removeResponse = await fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/remove-label`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        task_id: taskID,
                        label_id: labelID
                    })
                });

                if (!removeResponse.ok) {
                    throw new Error(`Failed to remove label ${labelID}`);
                }
            }

            // Refresh the widget content
            const refreshResponse = await fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/refresh`);
            
            if (!refreshResponse.ok) {
                throw new Error('Failed to refresh widget');
            }
            
            const newHTML = await refreshResponse.text();
            
            // Find the widget container and replace its content
            const widgetContainer = document.querySelector(`.widget-type-vikunja`);
            if (widgetContainer) {
                // Create a temporary container to parse the new HTML
                const temp = document.createElement('div');
                temp.innerHTML = newHTML;
                const newWidget = temp.firstElementChild;
                
                // Replace the widget
                widgetContainer.replaceWith(newWidget);
                
                // Reinitialize the widget
                initVikunjaWidget(newWidget);
            }

            closeModal();
        } catch (error) {
            console.error('Error updating task:', error);
            alert('Nie udało się zaaktualizować zadania: ' + error.message);
        }
    }

    closeBtn.addEventListener('click', closeModal);
    cancelBtn.addEventListener('click', closeModal);
    saveBtn.addEventListener('click', saveTask);

    // Close modal when clicking outside
    modal.addEventListener('click', function(e) {
        if (e.target === modal) {
            closeModal();
        }
    });
}

function openCreateModal(widgetID) {
    const modal = document.getElementById('vikunja-create-modal');
    const titleInput = document.getElementById('vikunja-create-title');
    const dueDateInput = document.getElementById('vikunja-create-due-date');
    const labelsContainer = document.getElementById('vikunja-create-labels-container');

    // Clear the form
    titleInput.value = '';
    dueDateInput.value = '';

    // Fetch and display labels
    labelsContainer.innerHTML = '<p>Ładowanie etykiet...</p>';
    
    fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/labels`)
        .then(response => response.json())
        .then(labels => {
            labelsContainer.innerHTML = '';
            
            if (labels && labels.length > 0) {
                labels.forEach(label => {
                    const labelCheckbox = document.createElement('label');
                    labelCheckbox.className = 'vikunja-label-option';
                    
                    const color = label.hex_color && label.hex_color[0] !== '#' 
                        ? '#' + label.hex_color 
                        : label.hex_color || '#666';
                    
                    labelCheckbox.innerHTML = `
                        <input type="checkbox" value="${label.id}">
                        <span class="label" style="border-color: ${color}; color: ${color};">${label.title}</span>
                    `;
                    
                    labelsContainer.appendChild(labelCheckbox);
                });
            } else {
                labelsContainer.innerHTML = '<p>Brak dostępnych etykiet</p>';
            }
        })
        .catch(error => {
            console.error('Error fetching labels:', error);
            labelsContainer.innerHTML = '<p>Nie udało się załadować etykiet</p>';
        });

    modal.style.display = 'flex';

    // Handle close button
    const closeBtn = modal.querySelector('.vikunja-modal-close');
    const cancelBtn = modal.querySelector('.vikunja-btn-cancel');
    const createBtn = modal.querySelector('.vikunja-btn-create');

    function closeModal() {
        modal.style.display = 'none';
        // Remove event listeners
        closeBtn.removeEventListener('click', closeModal);
        cancelBtn.removeEventListener('click', closeModal);
        createBtn.removeEventListener('click', createTask);
    }

    async function createTask() {
        const title = titleInput.value.trim();
        const dueDate = dueDateInput.value;
        
        // Get selected label IDs
        const selectedLabels = Array.from(labelsContainer.querySelectorAll('input[type="checkbox"]:checked'))
            .map(checkbox => parseInt(checkbox.value));

        if (!title) {
            alert('Tytuł zadania nie może być pusty');
            return;
        }

        // Convert datetime-local format to RFC3339
        let formattedDueDate = '';
        if (dueDate) {
            const date = new Date(dueDate);
            formattedDueDate = date.toISOString();
        }

        try {
            // Create the task
            const createResponse = await fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/create-task`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    title: title,
                    due_date: formattedDueDate,
                    label_ids: selectedLabels
                })
            });

            if (!createResponse.ok) {
                throw new Error('Failed to create task');
            }

            const createdTask = await createResponse.json();

            // Refresh the widget content
            const refreshResponse = await fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/refresh`);
            
            if (!refreshResponse.ok) {
                throw new Error('Failed to refresh widget');
            }
            
            const newHTML = await refreshResponse.text();
            
            // Find the widget container and replace its content
            const widgetContainer = document.querySelector(`.widget-type-vikunja`);
            if (widgetContainer) {
                // Create a temporary container to parse the new HTML
                const temp = document.createElement('div');
                temp.innerHTML = newHTML;
                const newWidget = temp.firstElementChild;
                
                // Replace the widget
                widgetContainer.replaceWith(newWidget);
                
                // Reinitialize the widget
                initVikunjaWidget(newWidget);
            }

            closeModal();
        } catch (error) {
            console.error('Error creating task:', error);
            alert('Nie udało się utworzyć zadania: ' + error.message);
        }
    }

    closeBtn.addEventListener('click', closeModal);
    cancelBtn.addEventListener('click', closeModal);
    createBtn.addEventListener('click', createTask);

    // Close modal when clicking outside
    modal.addEventListener('click', function(e) {
        if (e.target === modal) {
            closeModal();
        }
    });
}

export default initVikunjaWidget;
