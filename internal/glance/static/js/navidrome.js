export default function(element) {
    const widget = new NavidromePlayer(element);
    widget.initialize();
}

class NavidromePlayer {
    constructor(element) {
        this.element = element;
        this.playerContainer = element.querySelector('.navidrome-player');
        this.url = this.playerContainer.dataset.url;
        this.authParams = this.playerContainer.dataset.auth;
        
        this.audio = element.querySelector('#navidromeAudio');
        this.playPauseBtn = element.querySelector('#navidromePlayPauseBtn');
        this.playIcon = element.querySelector('#navidromePlayIcon');
        this.pauseIcon = element.querySelector('#navidromePauseIcon');
        this.prevBtn = element.querySelector('#navidromePrevBtn');
        this.nextBtn = element.querySelector('#navidromeNextBtn');
        this.shuffleBtn = element.querySelector('#navidromeShuffleBtn');
        this.loopBtn = element.querySelector('#navidromeLoopBtn');
        this.playlistSelect = element.querySelector('#navidromePlaylistSelect');
        this.volumeSlider = element.querySelector('#navidromeVolume');
        
        this.coverImg = element.querySelector('#navidromeCover');
        this.coverPlaceholder = element.querySelector('#navidromeCoverPlaceholder');
        this.titleEl = element.querySelector('#navidromeTitle');
        this.artistEl = element.querySelector('#navidromeArtist');
        this.bgEl = element.querySelector('.navidrome-bg');
        
        this.playlist = [];
        this.originalPlaylist = [];
        this.currentIndex = -1;
        this.isShuffle = false;
        this.isLoop = false;
        
        this.bindEvents();
    }
    
    initialize() {
        // Check if we have playlists
        if (this.playlistSelect.options.length <= 1) {
            this.titleEl.textContent = "Brak playlist";
        } else {
            this.loadState();
        }
    }
    
    bindEvents() {
        this.playlistSelect.addEventListener('change', (e) => this.loadPlaylist(e.target.value));
        
        this.playPauseBtn.addEventListener('click', () => this.togglePlay());
        this.prevBtn.addEventListener('click', () => this.playPrev());
        this.nextBtn.addEventListener('click', () => this.playNext());
        this.shuffleBtn.addEventListener('click', () => this.toggleShuffle());
        this.loopBtn.addEventListener('click', () => this.toggleLoop());
        this.volumeSlider.addEventListener('input', (e) => this.setVolume(e.target.value));
        
        this.audio.addEventListener('ended', () => this.playNext(true));
        this.audio.addEventListener('play', () => this.updatePlayState(true));
        this.audio.addEventListener('pause', () => this.updatePlayState(false));
        this.audio.addEventListener('error', (e) => console.error('Audio error', e));
    }

    loadState() {
        const savedState = localStorage.getItem('navidromeState');
        if (savedState) {
            try {
                const state = JSON.parse(savedState);
                
                if (state.volume !== undefined) {
                    this.setVolume(state.volume);
                    this.volumeSlider.value = state.volume;
                }

                if (state.isShuffle) {
                    this.isShuffle = true;
                    this.shuffleBtn.classList.add('active');
                }

                if (state.isLoop) {
                    this.isLoop = true;
                    this.loopBtn.classList.add('active');
                }

                if (state.playlistId) {
                    this.playlistSelect.value = state.playlistId;
                    // Load playlist but don't auto-play unless it was playing? 
                    // For now just load and set index
                    this.loadPlaylist(state.playlistId, state.songId);
                }
            } catch (e) {
                console.error('Failed to load navidrome state', e);
            }
        }
    }

    saveState() {
        const currentSong = this.playlist[this.currentIndex];
        const state = {
            playlistId: this.playlistSelect.value,
            songId: currentSong ? currentSong.id : null,
            isShuffle: this.isShuffle,
            isLoop: this.isLoop,
            volume: this.volumeSlider.value
        };
        localStorage.setItem('navidromeState', JSON.stringify(state));
    }
    
    async loadPlaylist(playlistId, initialSongId = null) {
        try {
            const response = await fetch(`${this.url}/rest/getPlaylist?${this.authParams}&id=${playlistId}`);
            const data = await response.json();
            
            if (data['subsonic-response'] && data['subsonic-response'].playlist && data['subsonic-response'].playlist.entry) {
                this.originalPlaylist = data['subsonic-response'].playlist.entry;
                this.playlist = [...this.originalPlaylist];
                
                if (this.isShuffle) {
                    this.shufflePlaylist();
                }
                
                this.currentIndex = 0;
                if (initialSongId) {
                    const index = this.playlist.findIndex(s => s.id === initialSongId);
                    if (index !== -1) {
                        this.currentIndex = index;
                    }
                }

                // Don't auto-play on load, just set info
                const song = this.playlist[this.currentIndex];
                this.updateInfo(song);
                
                // Pre-set audio src but don't play
                const streamUrl = `${this.url}/rest/stream?${this.authParams}&id=${song.id}`;
                this.audio.src = streamUrl;

                this.saveState();
            }
        } catch (error) {
            console.error('Failed to load playlist', error);
        }
    }

    updateInfo(song) {
        if (!song) return;
        
        this.titleEl.textContent = song.title;
        this.artistEl.textContent = song.artist;
        
        if (song.coverArt) {
            const coverUrl = `${this.url}/rest/getCoverArt?${this.authParams}&id=${song.coverArt}&size=300`;
            this.coverImg.src = coverUrl;
            this.coverImg.style.display = 'block';
            this.coverPlaceholder.style.display = 'none';
            this.bgEl.style.backgroundImage = `url('${coverUrl}')`;
        } else {
            this.coverImg.style.display = 'none';
            this.coverPlaceholder.style.display = 'flex';
            this.bgEl.style.backgroundImage = 'none';
        }
    }
    
    playSong(index) {
        if (index < 0 || index >= this.playlist.length) return;
        
        this.currentIndex = index;
        const song = this.playlist[index];
        
        this.updateInfo(song);
        
        // Stream URL
        const streamUrl = `${this.url}/rest/stream?${this.authParams}&id=${song.id}`;
        this.audio.src = streamUrl;
        this.audio.play().catch(e => console.error("Play failed", e));
        
        this.saveState();
    }
    
    togglePlay() {
        if (this.audio.paused) {
            if (this.currentIndex === -1 && this.playlist.length > 0) {
                this.playSong(0);
            } else {
                this.audio.play();
            }
        } else {
            this.audio.pause();
        }
    }
    
    updatePlayState(isPlaying) {
        if (isPlaying) {
            this.playIcon.style.display = 'none';
            this.pauseIcon.style.display = 'block';
        } else {
            this.playIcon.style.display = 'block';
            this.pauseIcon.style.display = 'none';
        }
    }
    
    playNext(auto = false) {
        if (this.playlist.length === 0) return;
        
        let nextIndex = this.currentIndex + 1;
        
        if (nextIndex >= this.playlist.length) {
            if (this.isLoop) {
                nextIndex = 0;
            } else {
                // Stop if not looping and auto-advanced
                if (auto) return;
                nextIndex = 0; // Manual next goes to start? Or stops? Let's go to start.
            }
        }
        
        this.playSong(nextIndex);
    }
    
    playPrev() {
        if (this.playlist.length === 0) return;
        let prevIndex = this.currentIndex - 1;
        if (prevIndex < 0) prevIndex = this.playlist.length - 1; // Always loop on manual prev
        this.playSong(prevIndex);
    }
    
    toggleShuffle() {
        this.isShuffle = !this.isShuffle;
        this.shuffleBtn.classList.toggle('active', this.isShuffle);
        
        const currentSong = this.playlist[this.currentIndex];
        
        if (this.isShuffle) {
            this.shufflePlaylist();
            // Find current song in new playlist to keep playing it
            if (currentSong) {
                this.currentIndex = this.playlist.findIndex(s => s.id === currentSong.id);
            }
        } else {
            this.playlist = [...this.originalPlaylist];
            if (currentSong) {
                this.currentIndex = this.playlist.findIndex(s => s.id === currentSong.id);
            }
        }
        this.saveState();
    }

    toggleLoop() {
        this.isLoop = !this.isLoop;
        this.loopBtn.classList.toggle('active', this.isLoop);
        this.saveState();
    }

    setVolume(value) {
        this.audio.volume = value / 100;
        this.saveState();
    }
    
    shufflePlaylist() {
        for (let i = this.playlist.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [this.playlist[i], this.playlist[j]] = [this.playlist[j], this.playlist[i]];
        }
    }
}
