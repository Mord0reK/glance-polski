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
        this.currentStationName = element.querySelector('#radyjkoCurrentName');
        this.currentStationIcon = element.querySelector('#radyjkoStationImg');
        
        // Get stations from hidden data container
        const stationDataElements = element.querySelectorAll('#radyjkoStationsData .radyjko-station-item');
        this.stations = Array.from(stationDataElements).map((el) => ({
            index: parseInt(el.dataset.index),
            url: el.dataset.url,
            name: el.dataset.name,
            icon: el.dataset.icon,
            shortName: el.dataset.shortname
        }));
        
        this.currentStationIndex = 0;
        this.isPlaying = false;
        this.hls = null;
    }

    initialize() {
        console.log('Initializing Radyjko player with', this.stations.length, 'stations');
        
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

        // Audio event listeners
        this.audioPlayer.addEventListener('play', () => this.onAudioPlay());
        this.audioPlayer.addEventListener('pause', () => this.onAudioPause());
        this.audioPlayer.addEventListener('error', () => this.onAudioError());
        this.audioPlayer.addEventListener('canplay', () => this.onCanPlay());
        
        // Load first station
        this.loadStation(0);
    }

    onCanPlay() {
        console.log('Audio can play');
    }

    previousStation() {
        if (this.currentStationIndex > 0) {
            this.loadStation(this.currentStationIndex - 1);
        }
    }

    nextStation() {
        if (this.currentStationIndex < this.stations.length - 1) {
            this.loadStation(this.currentStationIndex + 1);
        }
    }

    loadStation(index) {
        if (index < 0 || index >= this.stations.length) {
            console.error('Invalid station index:', index);
            return;
        }
        
        this.currentStationIndex = index;
        const station = this.stations[index];
        
        console.log('Loading station:', station);

        // Update UI
        this.currentStationName.textContent = station.name;
        
        // Update icon
        if (this.currentStationIcon) {
            this.currentStationIcon.src = station.icon;
            this.currentStationIcon.alt = station.name;
            this.currentStationIcon.style.display = 'block';
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

        // Pause current playback
        this.audioPlayer.pause();
        
        this.loadStationStream(station.url);
    }

    loadStationStream(streamUrl) {
        if (!streamUrl) {
            console.error('No stream URL provided');
            return;
        }

        console.log('Stream URL:', streamUrl);

        // Check if it's an HLS stream (m3u8)
        if (streamUrl.includes('.m3u8') && typeof Hls !== 'undefined' && Hls.isSupported()) {
            console.log('Using HLS for stream');
            this.hls = new Hls({
                enableWorker: false,
                lowLatencyMode: true,
            });
            
            this.hls.loadSource(streamUrl);
            this.hls.attachMedia(this.audioPlayer);
            
            this.hls.on(Hls.Events.MANIFEST_PARSED, () => {
                console.log('HLS manifest parsed, levels:', this.hls.levels.length);
                if (this.isPlaying) {
                    this.audioPlayer.play().catch(e => {
                        console.error('Error playing audio after HLS:', e);
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
        } else if (streamUrl.includes('.aac') || streamUrl.includes('timeradio-p')) {
            // Handle AAC streams
            console.log('Using AAC stream');
            this.audioPlayer.src = streamUrl;
            this.audioPlayer.type = 'audio/aac';
            if (this.isPlaying) {
                this.audioPlayer.play().catch(e => {
                    console.error('Error playing AAC audio:', e);
                });
            }
        } else {
            // Other formats
            console.log('Using generic audio stream');
            this.audioPlayer.src = streamUrl;
            if (this.isPlaying) {
                this.audioPlayer.play().catch(e => {
                    console.error('Error playing audio:', e);
                });
            }
        }
    }

    togglePlayPause() {
        console.log('Toggle play/pause, currently playing:', this.isPlaying);
        if (this.isPlaying) {
            this.pause();
        } else {
            this.play();
        }
    }

    play() {
        console.log('Play requested');
        const playPromise = this.audioPlayer.play();
        if (playPromise !== undefined) {
            playPromise
                .then(() => {
                    console.log('Audio playback started');
                })
                .catch(error => {
                    console.error('Error playing audio:', error);
                });
        }
    }

    pause() {
        console.log('Pause requested');
        this.audioPlayer.pause();
    }

    setVolume(value) {
        this.audioPlayer.volume = Math.max(0, Math.min(1, value));
        console.log('Volume set to:', this.audioPlayer.volume);
    }

    onAudioPlay() {
        console.log('Audio play event fired');
        this.isPlaying = true;
        this.updatePlayButton();
    }

    onAudioPause() {
        console.log('Audio pause event fired');
        this.isPlaying = false;
        this.updatePlayButton();
    }

    onAudioError() {
        console.error('Audio error:', this.audioPlayer.error?.message);
        this.isPlaying = false;
        this.updatePlayButton();
    }

    updatePlayButton() {
        const icon = this.isPlaying 
            ? '<svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 24 24"><path d="M6 4h4v16H6V4zm8 0h4v16h-4V4z"/></svg>'
            : '<svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>';
        
        this.playPauseBtn.innerHTML = icon;
    }


}
