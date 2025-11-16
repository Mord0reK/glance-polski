// Vikunja widget interactivity
export default function(widget) {
    if (!widget) return;
    
    {
        const widgetID = widget.querySelector('.vikunja-table')?.dataset.widgetId;
        if (!widgetID) return;

        // Handle task completion checkboxes
        widget.querySelectorAll('.vikunja-task-checkbox').forEach(checkbox => {
            checkbox.addEventListener('change', function(e) {
                if (!this.checked) {
                    this.checked = false;
                    return;
                }

                const row = this.closest('tr');
                const taskID = parseInt(row.dataset.taskId);

                if (!confirm('Czy na pewno chcesz oznaczyć to zadanie jako wykonane?')) {
                    this.checked = false;
                    return;
                }

                // Call API to complete task
                fetch(`/api/vikunja/${widgetID}/complete-task`, {
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
                    // Remove the row with animation
                    row.style.transition = 'opacity 0.3s ease';
                    row.style.opacity = '0';
                    setTimeout(() => {
                        row.remove();
                        // Check if there are no more tasks
                        const tbody = widget.querySelector('.vikunja-table tbody');
                        if (!tbody || tbody.children.length === 0) {
                            const table = widget.querySelector('.vikunja-table');
                            if (table) {
                                table.innerHTML = '<div class="flex items-center justify-center padding-block-5"><p>Brak zadań do wykonania</p></div>';
                            }
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
    
    fetch(`/api/vikunja/${widgetID}/labels`)
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

    function saveTask() {
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

        // Call API to update task
        fetch(`/api/vikunja/${widgetID}/update-task`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                task_id: taskID,
                title: newTitle,
                due_date: formattedDueDate,
                label_ids: selectedLabels
            })
        })
        .then(response => {
            if (!response.ok) throw new Error('Failed to update task');
            return response.json();
        })
        .then(data => {
            // Update the row with new data
            const titleCell = row.querySelector('.vikunja-title');
            if (titleCell) {
                titleCell.textContent = newTitle;
            }

            // Update the edit button's data attributes
            const editBtn = row.querySelector('.vikunja-edit-btn');
            if (editBtn) {
                editBtn.dataset.taskTitle = newTitle;
                if (newDueDate) {
                    editBtn.dataset.taskDueDate = newDueDate.replace('T', ' ');
                }
            }

            // Update labels (simplified - would need full refresh for accurate display)
            const labelsCell = row.querySelector('.vikunja-labels');
            if (labelsCell && selectedLabels.length > 0) {
                const labelContainer = labelsCell.querySelector('.label-container');
                if (labelContainer) {
                    // For now, just show that labels were updated
                    // A full page refresh would show the actual labels
                    alert('Zadanie zostało zaktualizowane. Odśwież stronę, aby zobaczyć wszystkie zmiany.');
                }
            }

            closeModal();
        })
        .catch(error => {
            console.error('Error updating task:', error);
            alert('Nie udało się zaktualizować zadania');
        });
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
