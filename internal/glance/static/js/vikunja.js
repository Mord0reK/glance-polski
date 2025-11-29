// Vikunja widget interactivity
let flatpickrLoadingPromise = null;


// Function to handle input color change based on content
function handleInputColorChange(input) {
    if (input.value.trim() !== '') {
        input.classList.add('has-value');
    } else {
        input.classList.remove('has-value');
    }
}

function loadFlatpickr() {
    if (window.flatpickr) return Promise.resolve();
    if (flatpickrLoadingPromise) return flatpickrLoadingPromise;


    flatpickrLoadingPromise = new Promise(async (resolve, reject) => {
        try {
            // Load CSS
            const cssLink = document.createElement("link");
            cssLink.rel = "stylesheet";
            cssLink.href = "https://cdn.jsdelivr.net/npm/flatpickr/dist/flatpickr.min.css";
            document.head.appendChild(cssLink);


            // Load JS
            await new Promise((res, rej) => {
                const script = document.createElement("script");
                script.src = "https://cdn.jsdelivr.net/npm/flatpickr";
                script.onload = res;
                script.onerror = rej;
                document.head.appendChild(script);
            });


            // Load Locale
            await new Promise((res, rej) => {
                const script = document.createElement("script");
                script.src = "https://cdn.jsdelivr.net/npm/flatpickr/dist/l10n/pl.js";
                script.onload = res;
                script.onerror = rej;
                document.head.appendChild(script);
            });


            resolve();
        } catch (e) {
            console.error("Failed to load flatpickr", e);
            reject(e);
        }
    });


    return flatpickrLoadingPromise;
}


function initVikunjaWidget(widget) {
    if (!widget) return;
    
    // Start loading flatpickr immediately
    loadFlatpickr();
    
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


                // Play completion sound
                const completionSound = document.getElementById('vikunja-completion-sound');
                if (completionSound) {
                    completionSound.currentTime = 0;
                    completionSound.play().catch(err => {
                        console.log('Could not play completion sound:', err);
                    });
                }


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
                const taskReminderDate = this.dataset.taskReminderDate;
                const affineNoteURL = this.dataset.affineNoteUrl || '';
                const customLinkURL = this.dataset.customLinkUrl || '';
                const customLinkTitle = this.dataset.customLinkTitle || '';
                
                // Get current labels
                const currentLabels = Array.from(row.querySelectorAll('.label')).map(label => 
                    parseInt(label.dataset.labelId)
                );


                openEditModal(widgetID, taskID, taskTitle, taskDueDate, taskReminderDate, currentLabels, row, affineNoteURL, customLinkURL, customLinkTitle);
            });
        });
    }
}


function initFlatpickr(element, defaultDate) {
    if (element._flatpickr) {
        element._flatpickr.destroy();
    }
    
    if (!window.flatpickr) {
        console.error("Flatpickr not loaded yet!");
        return;
    }
    
    // Find the wrapper element (parent of the input)
    const wrapper = element.closest('.flatpickr-wrapper');
    
    const fp = flatpickr(wrapper || element, {
        wrap: true, // Enable wrap mode to use external toggle
        enableTime: true,
        dateFormat: "Y-m-d H:i",
        time_24hr: true,
        locale: "pl",
        defaultDate: defaultDate || null,
        disableMobile: true,
        allowInput: true, // Allow manual input
        clickOpens: false // Only open on toggle button click
    });


    // Ensure the input element has reference to the flatpickr instance
    // This is needed because we initialize on the wrapper but access it via the input in other functions
    if (wrapper && element !== wrapper) {
        element._flatpickr = fp;
    }


    // Auto-format input as YYYY-MM-DD HH:MM
    element.addEventListener('input', function(e) {
        // Don't format if deleting content to allow easier editing
        if (e.inputType && e.inputType.startsWith('delete')) {
            return;
        }


        let v = this.value.replace(/\D/g, '');
        if (v.length > 12) v = v.slice(0, 12);
        
        let formatted = '';
        if (v.length > 0) formatted += v.slice(0, 4);
        if (v.length >= 4) formatted += '-';
        if (v.length > 4) formatted += v.slice(4, 6);
        if (v.length >= 6) formatted += '-';
        if (v.length > 6) formatted += v.slice(6, 8);
        if (v.length >= 8) formatted += ' ';
        if (v.length > 8) formatted += v.slice(8, 10);
        if (v.length >= 10) formatted += ':';
        if (v.length > 10) formatted += v.slice(10, 12);
        
        if (this.value !== formatted) {
            this.value = formatted;
        }
    });


    return fp;
}


async function openEditModal(widgetID, taskID, title, dueDate, reminderDate, currentLabelIDs, row, affineNoteURL, customLinkURL, customLinkTitle) {
    await loadFlatpickr();


    const modal = document.getElementById('vikunja-edit-modal');
    const titleInput = document.getElementById('vikunja-edit-title');
    const dueDateInput = document.getElementById('vikunja-edit-due-date');
    const affineURLInput = document.getElementById('vikunja-edit-affine-url');
    const customLinkURLInput = document.getElementById('vikunja-edit-custom-link-url');
    const customLinkTitleInput = document.getElementById('vikunja-edit-custom-link-title');
    const labelsContainer = document.getElementById('vikunja-labels-container');


    // Set current values
    titleInput.value = title || '';
    affineURLInput.value = affineNoteURL || '';
    customLinkURLInput.value = customLinkURL || '';
    customLinkTitleInput.value = customLinkTitle || '';
    
    // Set up input color change handlers
    handleInputColorChange(titleInput);
    titleInput.addEventListener('input', function() {
        handleInputColorChange(this);
    });
    
    handleInputColorChange(affineURLInput);
    affineURLInput.addEventListener('input', function() {
        handleInputColorChange(this);
    });
    
    handleInputColorChange(customLinkURLInput);
    customLinkURLInput.addEventListener('input', function() {
        handleInputColorChange(this);
    });
    
    handleInputColorChange(customLinkTitleInput);
    customLinkTitleInput.addEventListener('input', function() {
        handleInputColorChange(this);
    });
    
    // Initialize flatpickr
    initFlatpickr(dueDateInput, dueDate);
    handleInputColorChange(dueDateInput);
    dueDateInput.addEventListener('input', function() {
        handleInputColorChange(this);
    });


    // Fetch and display labels
    labelsContainer.innerHTML = '<p>Ładowanie etykiet...</p>';
    
    fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/labels`)
        .then(response => response.json())
        .then(labels => {
            labelsContainer.innerHTML = '';
            
            if (labels && labels.length > 0) {
                labels.forEach(label => {
                    const labelOption = document.createElement('div');
                    labelOption.className = 'vikunja-label-option';
                    
                    const color = label.hex_color && label.hex_color[0] !== '#' 
                        ? '#' + label.hex_color 
                        : label.hex_color || '#666';
                    
                    const isChecked = currentLabelIDs.includes(label.id);
                    
                    const checkbox = document.createElement('input');
                    checkbox.type = 'checkbox';
                    checkbox.value = label.id;
                    if (isChecked) checkbox.checked = true;

                    const labelText = document.createElement('span');
                    labelText.className = 'label';
                    labelText.style.borderColor = color;
                    labelText.style.color = color;
                    labelText.textContent = label.title;

                    labelOption.appendChild(checkbox);
                    labelOption.appendChild(labelText);
                    labelsContainer.appendChild(labelOption);
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
        const dueDateFP = dueDateInput._flatpickr.selectedDates[0];
        const affineNoteURL = affineURLInput.value.trim();
        const customLinkURL = customLinkURLInput.value.trim();
        const customLinkTitle = customLinkTitleInput.value.trim();
        
        // Get selected label IDs
        const selectedLabels = Array.from(labelsContainer.querySelectorAll('input[type="checkbox"]:checked'))
            .map(checkbox => parseInt(checkbox.value));


        if (!newTitle) {
            alert('Tytuł zadania nie może być pusty');
            return;
        }


        // Convert to RFC3339
        let formattedDueDate = '';
        if (dueDateFP) {
            formattedDueDate = dueDateFP.toISOString();
        }


        try {
            // Step 1: Update title, due date, Affine note URL, custom link URL and custom link title
            const updateResponse = await fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/update-task`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    task_id: taskID,
                    title: newTitle,
                    due_date: formattedDueDate,
                    affine_note_url: affineNoteURL,
                    custom_link_url: customLinkURL,
                    custom_link_title: customLinkTitle
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


async function openCreateModal(widgetID) {
    await loadFlatpickr();


    const modal = document.getElementById('vikunja-create-modal');
    const titleInput = document.getElementById('vikunja-create-title');
    const dueDateInput = document.getElementById('vikunja-create-due-date');
    const affineURLInput = document.getElementById('vikunja-create-affine-url');
    const customLinkURLInput = document.getElementById('vikunja-create-custom-link-url');
    const customLinkTitleInput = document.getElementById('vikunja-create-custom-link-title');
    const projectSelect = document.getElementById('vikunja-create-project');
    const labelsContainer = document.getElementById('vikunja-create-labels-container');


    // Clear the form
    titleInput.value = '';
    affineURLInput.value = '';
    customLinkURLInput.value = '';
    customLinkTitleInput.value = '';
    
    // Set up input color change handlers
    handleInputColorChange(titleInput);
    titleInput.addEventListener('input', function() {
        handleInputColorChange(this);
    });
    
    handleInputColorChange(affineURLInput);
    affineURLInput.addEventListener('input', function() {
        handleInputColorChange(this);
    });
    
    handleInputColorChange(customLinkURLInput);
    customLinkURLInput.addEventListener('input', function() {
        handleInputColorChange(this);
    });
    
    handleInputColorChange(customLinkTitleInput);
    customLinkTitleInput.addEventListener('input', function() {
        handleInputColorChange(this);
    });
    
    initFlatpickr(dueDateInput);
    handleInputColorChange(dueDateInput);
    dueDateInput.addEventListener('input', function() {
        handleInputColorChange(this);
    });
    
    // Fetch and populate projects
    projectSelect.innerHTML = '<option value="">Ładowanie...</option>';
    
    fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/projects`)
        .then(response => response.json())
        .then(projects => {
            projectSelect.innerHTML = '<option value="">Domyślny projekt</option>';
            
            if (projects && projects.length > 0) {
                projects.forEach(project => {
                    const option = document.createElement('option');
                    option.value = project.id;
                    option.textContent = project.title;
                    projectSelect.appendChild(option);
                });
            }
        })
        .catch(error => {
            console.error('Error fetching projects:', error);
            projectSelect.innerHTML = '<option value="">Błąd ładowania projektów</option>';
        });


    // Fetch and display labels
    labelsContainer.innerHTML = '<p>Ładowanie etykiet...</p>';
    
    fetch(`${pageData.baseURL}/api/vikunja/${widgetID}/labels`)
        .then(response => response.json())
        .then(labels => {
            labelsContainer.innerHTML = '';
            
            if (labels && labels.length > 0) {
                labels.forEach(label => {
                    const labelOption = document.createElement('div');
                    labelOption.className = 'vikunja-label-option';
                    
                    const color = label.hex_color && label.hex_color[0] !== '#' 
                        ? '#' + label.hex_color 
                        : label.hex_color || '#666';
                    
                    const checkbox = document.createElement('input');
                    checkbox.type = 'checkbox';
                    checkbox.value = label.id;

                    const labelText = document.createElement('span');
                    labelText.className = 'label';
                    labelText.style.borderColor = color;
                    labelText.style.color = color;
                    labelText.textContent = label.title;

                    labelOption.appendChild(checkbox);
                    labelOption.appendChild(labelText);
                    labelsContainer.appendChild(labelOption);
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
        const dueDateFP = dueDateInput._flatpickr.selectedDates[0];
        const affineNoteURL = affineURLInput.value.trim();
        const customLinkURL = customLinkURLInput.value.trim();
        const customLinkTitle = customLinkTitleInput.value.trim();
        const projectID = projectSelect.value ? parseInt(projectSelect.value) : 0;
        
        // Get selected label IDs
        const selectedLabels = Array.from(labelsContainer.querySelectorAll('input[type="checkbox"]:checked'))
            .map(checkbox => parseInt(checkbox.value));


        if (!title) {
            alert('Tytuł zadania nie może być pusty');
            return;
        }


        // Convert datetime-local format to RFC3339
        let formattedDueDate = '';
        if (dueDateFP) {
            formattedDueDate = dueDateFP.toISOString();
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
                    label_ids: selectedLabels,
                    project_id: projectID,
                    affine_note_url: affineNoteURL,
                    custom_link_url: customLinkURL,
                    custom_link_title: customLinkTitle
                })
            });


            if (!createResponse.ok) {
                const errorText = await createResponse.text();
                console.error('Create task failed:', createResponse.status, errorText);
                throw new Error(`Błąd ${createResponse.status}: ${errorText || createResponse.statusText}`);
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
