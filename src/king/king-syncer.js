import got from "got";
import {to} from "await-to-js";

export class KingSyncer {

    #timer = -1;

    constructor(context) {
        this.context = context;
    }

    async #sync() {
        const [err, response] = await to(got(`${this.context.councilHost}/king`, {
            method: "PUT",
            json: {
                shutting_down: this.context.shuttingDown,
                ratholes: this.context.config.ratholes,
                host: this.context.host,
                location: this.context.location,
                ready_service_ids: this.context.readyServiceIds,
            },
        }));
        if (err || response.statusCode !== 200) {
            console.error("msg=\"Failed to sync with council\" service.type=ratking", err.message, response?.statusCode ?? 0);
        }
    }

    async stop() {
        clearTimeout(this.#timer);
        await this.#sync();
    }

    start() {
        this.#sync().then(() => {
            this.#timer = setTimeout(() => this.start(), 1000);
        });
    }
}
