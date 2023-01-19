import got from "got";
import {to} from "await-to-js";

export class LingSyncer {

    constructor(context) {
        this.context = context;
    }

    async #put() {
        const ratholes = Array.from(this.context.config.ratholeMap.values());
        const readyServiceIds = this.context.readyServiceIds;
        const [err, response] = await to(got(`${this.context.councilHost}/ling`, {
            method: "PUT",
            json: {ling_id: this.context.lingId, ratholes, ready_service_ids: readyServiceIds, prefered_location: "mylocation"}, // TODO: From cli options
        }));
        if (err || response.statusCode !== 200) {
            console.error("msg=\"Failed to sync with council\" service_type=ratling", err.message, response?.statusCode ?? 0);
        }
    }

    start() {
        this.#put().then(() => {
            setTimeout(() => this.start(), 1000);
        });
    }
}
