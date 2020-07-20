import { Controller } from 'stimulus'

export default class extends Controller {
    static targets = ['episode']

    connect() {
        window.addEventListener('episode:update', this.update)
        console.log(this.episodeTargets)
    }

    disconnect() {
        window.removeEventListener('episode:update', this.update)
    }

    update = (e) => {
        const { id, listened } = e.detail
        const updated = this.episodeTargets.find(ep => ep.dataset.id === id)

        if (!updated) return;

        if (listened) updated.classList.add('Calendar__day__episode--listened');
        else updated.classList.remove('Calendar__day__episode--listened');
    }
}