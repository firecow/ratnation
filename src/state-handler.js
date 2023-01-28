import got from "got";
import {to} from "await-to-js";

export class StateHandler {

    #state = null;
    #updatedFunc;
    #councilHost;
    #timer = -1;

    constructor({councilHost, updatedFunc}) {
        this.#updatedFunc = updatedFunc;
        this.#councilHost = councilHost;
    }

    hasState() {
        return this.#state !== null;
    }

    stop() {
        clearTimeout(this.#timer);
    }

    start() {
        this.fetchState().then(() => {
            this.#timer = setTimeout(() => this.start(), 500);
        });
    }

    #update(newState) {
        this.#state = newState;
        this.#updatedFunc(newState);
    }

    async fetchState() {
        const [err, response] = await to(got(`${this.#councilHost}/state`));
        if (err || response.statusCode !== 200) {
            return console.error("Failed to fetch state from council", err.message, response?.statusCode ?? 0);
        }

        const newState = JSON.parse(response.body);
        if (this.#state === null || this.#state["revision"] !== newState["revision"]) {
            this.#update(newState);
        }
    }
}
