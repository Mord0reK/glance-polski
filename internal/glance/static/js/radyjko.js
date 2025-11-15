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
            shortName: el.dataset.shortname,
            isOpenFM: el.dataset.isopenfm === '1',
            openFmId: el.dataset.openfmid ? parseInt(el.dataset.openfmid) : null
        }));
        
        this.currentStationIndex = 0;
        this.isPlaying = false;
        this.hls = null;
        this.hlsLoaded = false;
        this.hlsLoading = false;
    }

    async loadHlsJs() {
        if (this.hlsLoaded) {
            console.log('HLS.js already loaded');
            return true;
        }
        
        if (this.hlsLoading) {
            console.log('HLS.js loading in progress, waiting...');
            // Wait for existing load to complete
            return new Promise((resolve) => {
                const checkInterval = setInterval(() => {
                    if (this.hlsLoaded) {
                        clearInterval(checkInterval);
                        console.log('HLS.js finished loading');
                        resolve(true);
                    }
                }, 100);
            });
        }
        
        this.hlsLoading = true;
        console.log('Starting to load HLS.js...');
        
        return new Promise((resolve, reject) => {
            if (typeof window.Hls !== 'undefined') {
                console.log('HLS.js already available globally');
                this.hlsLoaded = true;
                this.hlsLoading = false;
                resolve(true);
                return;
            }
            
            const script = document.createElement('script');
            // Use specific stable version instead of @latest
            script.src = 'https://cdn.jsdelivr.net/npm/hls.js@1.5.15/dist/hls.min.js';
            script.onload = () => {
                console.log('HLS.js script loaded successfully, version:', window.Hls?.version);
                console.log('HLS.isSupported:', window.Hls?.isSupported());
                this.hlsLoaded = true;
                this.hlsLoading = false;
                resolve(true);
            };
            script.onerror = (error) => {
                console.error('Failed to load HLS.js script:', error);
                this.hlsLoading = false;
                reject(false);
            };
            document.head.appendChild(script);
            console.log('HLS.js script tag added to document head');
        });
    }

    initialize() {
        console.log('Initializing Radyjko player with', this.stations.length, 'stations');
        
        // Pre-load HLS.js if any station uses m3u8
        const hasM3u8Station = this.stations.some(s => s.url && s.url.includes('.m3u8'));
        if (hasM3u8Station) {
            console.log('Detected m3u8 stations, pre-loading HLS.js...');
            this.loadHlsJs().then(() => {
                console.log('HLS.js pre-loaded successfully');
            }).catch(err => {
                console.error('Failed to pre-load HLS.js:', err);
            });
        }
        
        // Set up play/pause button
        this.playPauseBtn.addEventListener('click', () => this.togglePlayPause());
        
        // Set up navigation buttons
        this.prevBtn.addEventListener('click', () => this.previousStation());
        this.nextBtn.addEventListener('click', () => this.nextStation());
        
        // Set up volume control
        this.volumeSlider.addEventListener('input', (e) => {
            this.setVolume(e.target.value / 100);
        });

        // Load saved volume or use default
        const savedVolume = localStorage.getItem('radyjko-volume');
        if (savedVolume !== null) {
            const volumeValue = parseFloat(savedVolume);
            this.volumeSlider.value = volumeValue * 100;
            this.setVolume(volumeValue);
            console.log('Loaded saved volume:', volumeValue);
        } else {
            this.setVolume(0.7);
        }

        // Audio event listeners
        this.audioPlayer.addEventListener('play', () => this.onAudioPlay());
        this.audioPlayer.addEventListener('pause', () => this.onAudioPause());
        this.audioPlayer.addEventListener('error', () => this.onAudioError());
        this.audioPlayer.addEventListener('canplay', () => this.onCanPlay());
        
        // Load last selected station or first station
        const savedStationIndex = localStorage.getItem('radyjko-last-station');
        if (savedStationIndex !== null) {
            const stationIndex = parseInt(savedStationIndex);
            if (stationIndex >= 0 && stationIndex < this.stations.length) {
                console.log('Loading last selected station:', stationIndex);
                this.loadStation(stationIndex);
            } else {
                this.loadStation(0);
            }
        } else {
            this.loadStation(0);
        }
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
        
        // Save last selected station to localStorage
        localStorage.setItem('radyjko-last-station', index);

        // Update UI
        this.currentStationName.textContent = station.name;
        
        // Update icon
        if (this.currentStationIcon) {
            this.currentStationIcon.src = station.icon;
            this.currentStationIcon.alt = station.name;
            this.currentStationIcon.style.display = 'block';
        }
        
        // Pause current playback
        this.audioPlayer.pause();
        
        // Destroy previous HLS instance if exists
        if (this.hls) {
            try {
                this.hls.destroy();
                console.log('Previous HLS instance destroyed');
            } catch (e) {
                console.error('Error destroying HLS:', e);
            }
            this.hls = null;
        }

        // Reset audio element completely
        this.audioPlayer.removeAttribute('src');
        this.audioPlayer.removeAttribute('type');
        this.audioPlayer.load();
        
        // Load stream with proper URL
        if (station.isOpenFM && station.openFmId) {
            this.loadOpenFMStation(station.openFmId);
        } else {
            this.loadStationStream(station.url);
        }
    }

    async loadOpenFMStation(stationId) {
        console.log('Loading OpenFM station:', stationId);
        try {
            // Fetch the actual stream URL from OpenFM API
            const apiUrl = `https://open.fm/api/user/token?fp=https://stream-cdn-1.open.fm/OFM${stationId}/ngrp:standard/playlist.m3u8`;
            const response = await fetch(apiUrl, { cache: 'no-store' });
            const data = await response.json();
            
            if (data && data.url) {
                console.log('OpenFM stream URL retrieved:', data.url);
                this.loadStationStream(data.url);
            } else {
                console.error('Failed to get OpenFM stream URL');
            }
        } catch (error) {
            console.error('Error fetching OpenFM stream:', error);
        }
    }

    async loadStationStream(streamUrl) {
        if (!streamUrl) {
            console.error('No stream URL provided');
            return;
        }

        console.log('Stream URL:', streamUrl);

        // Check if it's an HLS stream (m3u8)
        if (streamUrl.includes('.m3u8')) {
            // Try to load HLS.js
            try {
                const hlsLoadSuccess = await this.loadHlsJs();
                
                // Check if HLS.js is available and supported
                if (hlsLoadSuccess && typeof window.Hls !== 'undefined' && window.Hls.isSupported()) {
                    console.log('Using HLS.js for m3u8 stream');
                    this.hls = new window.Hls({
                        debug: false,
                        enableWorker: true,
                        lowLatencyMode: false,
                        backBufferLength: 90,
                        maxBufferLength: 30,
                        maxMaxBufferLength: 600,
                        maxBufferSize: 60 * 1000 * 1000,
                        maxBufferHole: 0.5,
                        manifestLoadingTimeOut: 10000,
                        manifestLoadingMaxRetry: 4,
                        manifestLoadingRetryDelay: 1000,
                        levelLoadingTimeOut: 10000,
                        levelLoadingMaxRetry: 4,
                        levelLoadingRetryDelay: 1000,
                        xhrSetup: function(xhr, url) {
                            xhr.withCredentials = false;
                        }
                    });
                    
                    this.hls.loadSource(streamUrl);
                    this.hls.attachMedia(this.audioPlayer);
                    
                    this.hls.on(window.Hls.Events.MANIFEST_PARSED, () => {
                        console.log('HLS manifest parsed successfully');
                        if (this.isPlaying) {
                            this.audioPlayer.play().catch(e => {
                                console.error('Error playing audio after HLS:', e);
                            });
                        }
                    });
                    
                    this.hls.on(window.Hls.Events.ERROR, (event, data) => {
                        console.error('HLS Error:', {
                            type: data.type,
                            details: data.details,
                            fatal: data.fatal
                        });
                        
                        if (data.fatal) {
                            switch(data.type) {
                                case window.Hls.ErrorTypes.NETWORK_ERROR:
                                    console.error('Fatal network error, trying to recover...');
                                    setTimeout(() => {
                                        if (this.hls) {
                                            this.hls.startLoad();
                                        }
                                    }, 1000);
                                    break;
                                case window.Hls.ErrorTypes.MEDIA_ERROR:
                                    console.error('Fatal media error, trying to recover...');
                                    if (this.hls) {
                                        this.hls.recoverMediaError();
                                    }
                                    break;
                                default:
                                    console.error('Fatal error, cannot recover');
                                    if (this.hls) {
                                        this.hls.destroy();
                                        this.hls = null;
                                    }
                                    this.isPlaying = false;
                                    this.updatePlayButton();
                                    break;
                            }
                        }
                    });
                } else if (this.audioPlayer.canPlayType('application/vnd.apple.mpegurl')) {
                    // Native HLS support (Safari)
                    console.log('Using native HLS support for m3u8 stream');
                    this.audioPlayer.src = streamUrl;
                    if (this.isPlaying) {
                        this.audioPlayer.play().catch(e => {
                            console.error('Error playing native HLS audio:', e);
                        });
                    }
                } else {
                    console.error('HLS not supported in this browser');
                }
            } catch (error) {
                console.error('Error loading HLS.js or setting up stream:', error);
            }
        } else if (streamUrl.includes('.aac') || streamUrl.includes('timeradio-p')) {
            // Handle AAC streams
            console.log('Using AAC stream');
            this.audioPlayer.src = streamUrl;
            this.audioPlayer.type = 'audio/aac';
            this.audioPlayer.load();
            if (this.isPlaying) {
                this.audioPlayer.play().catch(e => {
                    console.error('Error playing AAC audio:', e);
                });
            }
        } else if (streamUrl.includes('radiofreee') || streamUrl.includes('radio.lublin.pl')) {
            // Handle Radio Free stream - use proxy to avoid CORS issues
            console.log('Using Radio Free stream via proxy');
            const proxyUrl = `/api/audio-proxy?url=${encodeURIComponent(streamUrl)}`;
            this.audioPlayer.src = proxyUrl;
            this.audioPlayer.removeAttribute('type');
            this.audioPlayer.load();
            if (this.isPlaying) {
                this.audioPlayer.play().catch(e => {
                    console.error('Error playing Radio Free audio:', e);
                });
            }
        } else {
            // Other formats - let browser auto-detect
            console.log('Using generic audio stream');
            this.audioPlayer.src = streamUrl;
            this.audioPlayer.removeAttribute('type');
            this.audioPlayer.load();
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
        // Save volume to localStorage
        localStorage.setItem('radyjko-volume', this.audioPlayer.volume);
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
