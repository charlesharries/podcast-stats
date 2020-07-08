import { Controller } from 'stimulus';
import axios from 'axios';

export default class extends Controller {
    static targets = ['form', 'button', 'check']

    get subscribed() {
        return this.data.get('subscribed') === 'true'
    }

    set subscribed(subscribed) {
        return this.data.set('subscribed', subscribed)
    }

    async subscribe(e) {
        e.preventDefault();
        const collectionID = this.data.get('podcastId')

        const body = `collectionID=${encodeURIComponent(collectionID)}`

        const { data } = await axios.post(
            '/api/subscriptions',
            body,
            {
                withCredentials: true,
                headers: { 'Content-Type': 'application/x-www-form-urlencoded'}
            }
        );

        if (data.error) {
            console.error(data)
            return
        }

        this.subscribed = true
        this.updateUI();
    }

    async unsubscribe(e) {
        e.preventDefault();
        const collectionID = this.data.get('podcastId')

        const body = `collectionID=${encodeURIComponent(collectionID)}`
        
        const { data } = await axios.post(
            '/api/subscriptions/delete',
            body,
            {
                withCredentials: true,
                headers: { 'Content-Type': 'application/x-www-form-urlencoded'}
            }
        );

        if (data.error) {
            console.error(data)
            return
        }

        this.subscribed = false
        this.updateUI();
    }

    updateUI() {
        if (this.subscribed) {
            this.buttonTarget.innerText = 'Unsubscribe';
            this.checkTarget.style.display = 'inline-block';
            this.formTarget.dataset.action = 'subscription#unsubscribe'
        } else {
            this.buttonTarget.innerText = 'Subscribe';
            this.checkTarget.style.display = 'none';
            this.formTarget.dataset.action = 'subscription#subscribe'
        }
    }
}