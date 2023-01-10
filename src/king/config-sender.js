import got from "got";
import {to} from "await-to-js";

export class ConfigSender {

    constructor({councilHost, config, host, location}) {
        this.councilHost = councilHost;
        this.config = config;
        this.location = location;
        this.host = host;
    }

    async #sendConfig() {
        const ratholes = Array.from(this.config.ratholeMap.values());
        const [err, response] = await to(got(`${this.councilHost}/king`, {
            method: "PUT",
            json: {
                ratholes,
                host: this.host,
                location: this.location,
            },
        }));
        if (err || response.statusCode !== 200) {
            console.error("Failed to send config to council", err.message, response?.statusCode ?? 0);
        }
    }

    start() {
        this.#sendConfig().then(() => {
            setTimeout(() => this.start(), 5000);
        }).catch(err => console.error(err));
    }
}
