import axios from 'axios';
import { Controller} from 'stimulus';


export default class extends Controller {
    static targets = ['form', 'button']

    get listened() {
        return this.data.get('listened') === 'true'
    }

    set listened(listened) {
        return this.data.set('listened', listened)
    }

    updateUI() {
        if (this.listened) {
            this.element.classList.add('Episode--listened')
            this.buttonTarget.innerText = 'Unlisten';
            this.formTarget.dataset.action = this.formTarget.dataset.action.replace('#listen', '#unlisten')
        } else {
            this.element.classList.remove('Episode--listened');
            this.buttonTarget.innerText = 'Listen';
            this.formTarget.dataset.action = this.formTarget.dataset.action.replace('#unlisten', '#listen')
        }

        const updateEvent = new CustomEvent('episode:update', {
            detail: {
                id: this.data.get('id'),
                listened: this.listened,
            },
            bubbles: true,
        })
        this.formTarget.dispatchEvent(updateEvent)
    }

    async listen(e) {
        e.preventDefault();
        const episodeId = this.data.get('id');

        const { data } = await axios.post(`/api/episodes/${episodeId}/listens`, {}, {
            withCredentials: true,
        });

        if (data.error) {
            console.error(data)
            return
        }

        this.listened = true;
        this.updateUI()
    }

    async unlisten(e) {
        e.preventDefault();
        const episodeId = this.data.get('id');

        const { data } = await axios.post(`/api/episodes/${episodeId}/listens/delete`, {}, {
            withCredentials: true,
        });

        if (data.error) {
            console.error(data)
            return
        }

        this.listened = false;
        this.updateUI();
    }
}