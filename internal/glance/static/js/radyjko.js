export default function(element) {
    const widget = new RadyjkoPlayer(element);
    widget.initialize();
}

class RadyjkoPlayer {
    constructor(element) {
        this.element = element;
        this.audioPlayer = element.querySelector('#radyjkoAudioPlayer');
        this.playPauseBtn = element.querySelector('#radyjkoPlayPauseBtn');
        this.prevBtn = element.querySelector('#radyjkoPrevBtn');
        this.nextBtn = element.querySelector('#radyjkoNextBtn');
        this.volumeSlider = element.querySelector('#radyjkoVolume');
        this.stationsBtn = element.querySelector('#radyjkoStationsBtn');
        this.currentStationName = element.querySelector('#radyjkoCurrentName');
        this.currentStationIcon = element.querySelector('#radyjkoStationImg');
        this.stationItems = element.querySelectorAll('.radyjko-station-item');
        
        this.currentStationIndex = 0;
        this.isPlaying = false;
        this.hls = null;
    }

    initialize() {
        // Set up play/pause button
        this.playPauseBtn.addEventListener('click', () => this.togglePlayPause());
        
        // Set up navigation buttons
        this.prevBtn.addEventListener('click', () => this.previousStation());
        this.nextBtn.addEventListener('click', () => this.nextStation());
        
        // Set up volume control
        this.volumeSlider.addEventListener('input', (e) => {
            this.setVolume(e.target.value / 100);
        });

        // Set initial volume
        this.setVolume(0.7);

        // Set up station items in popover
        this.stationItems.forEach((item) => {
            item.addEventListener('click', () => {
                const index = parseInt(item.dataset.stationIndex);
                this.selectStation(index);
                if (!this.isPlaying) {
                    this.play();
                }
                // Close popover
                const popover = this.element.querySelector('[data-popover-html]');
                if (popover && popover.closest('.popover')) {
                    const closeBtn = popover.closest('.popover')?.querySelector('[data-popover-close]');
                    if (closeBtn) closeBtn.click();
                }
            });
        });

        // Audio event listeners
        this.audioPlayer.addEventListener('play', () => this.onAudioPlay());
        this.audioPlayer.addEventListener('pause', () => this.onAudioPause());
        this.audioPlayer.addEventListener('error', () => this.onAudioError());
        
        // Load first station by default
        this.selectStation(0);
    }

    previousStation() {
        if (this.currentStationIndex > 0) {
            this.selectStation(this.currentStationIndex - 1);
        }
    }

    nextStation() {
        if (this.currentStationIndex < this.stationItems.length - 1) {
            this.selectStation(this.currentStationIndex + 1);
        }
    }

    selectStation(index) {
        if (index < 0 || index >= this.stationItems.length) return;
        
        this.currentStationIndex = index;
        this.updateActiveStation();
        this.loadStation();
    }

    loadStation() {
        const station = this.getCurrentStation();
        if (!station) return;

        // Update UI
        const stationName = station.dataset.stationName;
        this.currentStationName.textContent = stationName;
        
        // Update icon
        const iconUrl = station.dataset.stationIcon;
        if (this.currentStationIcon) {
            this.currentStationIcon.src = iconUrl;
            this.currentStationIcon.alt = stationName;
            // Add error handler for images that fail to load
            this.currentStationIcon.onerror = () => {
                this.currentStationIcon.style.display = 'none';
            };
        }
        
        // Destroy previous HLS instance if exists
        if (this.hls) {
            try {
                this.hls.destroy();
            } catch (e) {
                console.error('Error destroying HLS:', e);
            }
            this.hls = null;
        }

        const streamUrl = station.dataset.stationUrl;

        // Check if it's an HLS stream (m3u8)
        if (streamUrl.includes('.m3u8') && typeof Hls !== 'undefined' && Hls.isSupported()) {
            this.hls = new Hls();
            this.hls.loadSource(streamUrl);
            this.hls.attachMedia(this.audioPlayer);
            
            this.hls.on(Hls.Events.MANIFEST_PARSED, () => {
                if (this.isPlaying) {
                    this.audioPlayer.play().catch(e => {
                        console.error('Error playing audio:', e);
                        this.updatePlayButton();
                    });
                }
            });

            this.hls.on(Hls.Events.ERROR, (event, data) => {
                console.error('HLS Error:', data);
                if (data.fatal) {
                    this.isPlaying = false;
                    this.updatePlayButton();
                }
            });
        } else {
            // For non-HLS streams or browsers without HLS support
            this.audioPlayer.src = streamUrl;
            
            if (this.isPlaying) {
                this.audioPlayer.play().catch(e => {
                    console.error('Error playing audio:', e);
                    this.updatePlayButton();
                });
            }
        }
    }

    togglePlayPause() {
        if (this.isPlaying) {
            this.pause();
        } else {
            this.play();
        }
    }

    play() {
        this.audioPlayer.play().catch(e => {
            console.error('Error playing audio:', e);
        });
    }

    pause() {
        this.audioPlayer.pause();
    }

    setVolume(value) {
        this.audioPlayer.volume = Math.max(0, Math.min(1, value));
    }

    onAudioPlay() {
        this.isPlaying = true;
        this.updatePlayButton();
    }

    onAudioPause() {
        this.isPlaying = false;
        this.updatePlayButton();
    }

    onAudioError() {
        this.isPlaying = false;
        this.updatePlayButton();
        console.error('Audio player error:', this.audioPlayer.error?.message);
    }

    updatePlayButton() {
        const icon = this.isPlaying 
            ? '<svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 24 24"><path d="M6 4h4v16H6V4zm8 0h4v16h-4V4z"/></svg>'
            : '<svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>';
        
        this.playPauseBtn.innerHTML = icon;
    }

    updateActiveStation() {
        this.stationItems.forEach((item, index) => {
            if (index === this.currentStationIndex) {
                item.classList.add('active');
            } else {
                item.classList.remove('active');
            }
        });
    }

    getCurrentStation() {
        return this.stationItems[this.currentStationIndex];
    }
}
