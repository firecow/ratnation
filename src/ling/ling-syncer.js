import got from "got";
import {to} from "await-to-js";

export class LingSyncer {

    #timer = -1;

    constructor(context) {
        this.context = context;
    }

    async #sync() {
        const [err, response] = await to(got(`${this.context.councilHost}/ling`, {
            method: "PUT",
            json: {
                ling_id: this.context.lingId,
                shutting_down: this.context.shuttingDown,
                ratholes: Array.from(this.context.config.ratholeMap.values()),
                ready_service_ids: this.context.readyServiceIds,
                prefered_location: "mylocation",
            },
        }));
        if (err || response.statusCode !== 200) {
            console.error("msg=\"Failed to sync with council\" service.type=ratling", err.message, response?.statusCode ?? 0);
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
