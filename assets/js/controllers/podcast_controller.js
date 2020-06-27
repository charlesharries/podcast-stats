import { Controller } from 'stimulus';

export default class extends Controller {
    static targets = ['episode', 'unlistenedTime', 'unlistenedEpisodes']

    connect() {
        console.log(this.episodeTargets)
    }

    update(e) {
        this.unlistenedEpisodesTarget.innerText = this.unlistenedEls().length
        this.unlistenedTimeTarget.innerText = this.unlistenedTime()
    }

    unlistenedEls() {
        return this.episodeTargets.filter(ep => {
            return ep.dataset.episodeListened !== 'true'
        })
    }

    unlistenedTime() {
        const secs = this.unlistenedEls().reduce((sum, el) => {
            return sum + parseInt(el.dataset.duration)
        }, 0)

        const h = Math.floor(secs / (60 * 60))
        const m = Math.floor((secs - (h * 60 * 60)) / 60)
    
        const hs = h > 0 ? `${h}h ` : ''
        const ms = m > 0 ? `${m}m ` : ''
    
        return `${hs}${ms}`;
    }
}