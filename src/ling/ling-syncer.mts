import got from "got";
import {to} from "await-to-js";
import {LingContext} from "./ling.mjs";
import {Ticker} from "../ticker.mjs";

export class LingSyncer extends Ticker {

    private readonly context;

    constructor (context: LingContext) {
        super({interval: 1000, tick: async () => this.#sync()});
        this.context = context;
    }

    async #sync () {
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
            console.error("msg=\"Failed to sync with council\" service.type=ratling", err?.message ?? response?.statusCode ?? 0);
        }
    }
}
