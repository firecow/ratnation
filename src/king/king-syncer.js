import got from "got";
import {to} from "await-to-js";

export class KingSyncer {

    constructor(context) {
        this.context = context;
    }

    async #put() {
        const [err, response] = await to(got(`${this.context.councilHost}/king`, {
            method: "PUT",
            json: {
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

    start() {
        this.#put().then(() => {
            setTimeout(() => this.start(), 1000);
        });
    }
}
