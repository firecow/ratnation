import got from "got";
import {to} from "await-to-js";
import {KingContext} from "./king.mjs";
import {Ticker} from "../ticker.mjs";

export class KingSyncer extends Ticker {

    private readonly context;

    constructor (context: KingContext) {
        super({interval: 1000, tick: async () => this.#sync()});
        this.context = context;
    }

    async #sync () {
        const [err, response] = await to(got(`${this.context.councilHost}/king`, {
            method: "PUT",
            json: {
                host: this.context.host,
                shutting_down: this.context.shuttingDown,
                ratholes: this.context.config.ratholes,
                ready_service_ids: this.context.readyServiceIds,
                location: this.context.location,
            },
        }));
        if (err || response.statusCode !== 200) {
            console.error("msg=\"Failed to sync with council\" service.type=ratking", err?.message ?? response?.statusCode ?? 0);
        }
    }

}
