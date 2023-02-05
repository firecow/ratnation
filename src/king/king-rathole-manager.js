import fs from "fs";
import {ProcessManager} from "../process-manager.js";

export class KingRatholeManager extends ProcessManager {

    #context;

    constructor(context) {
        super("king-rathole-manager");
        this.#context = context;
    }

    #getServices({bindPort, host}) {
        const state = this.#context.state;
        return state["services"].filter(s => {
            if (s["bind_port"] !== bindPort || s["host"] !== host) {
                return false;
            }

            const ling = state.lings.find(l => l["ling_id"] === s["ling_id"]);
            if (!ling) {
                console.error(`msg="Ling not found for service" service.name=${s["name"]} service.type=ratking log.logger=rathole-manager`);
                return false;
            }

            return !ling["shutting_down"];
        });
    }

    #writeRatholeFile({bindPort, services}) {
        const lines = [];
        lines.push(
            "[server]",
            `bind_addr = "0.0.0.0:${bindPort}"`,
            "",
            "[server.services]",
        );

        for (const service of services) {
            lines.push(
                `[server.services.${service["service_id"].replace(/:/g, "-")}]`,
                `token = "${service["token"]}"`,
                `bind_addr = "0.0.0.0:${service["remote_port"]}"`,
            );
        }

        const ratholeFile = `rathole-server-${bindPort}.toml`;
        fs.writeFileSync(`src/king/${ratholeFile}`, `${lines.join("\n")}\n`, "utf8");
        return ratholeFile;
    }

    async #each(ratholeCnf) {
        const bindPort = ratholeCnf.bind_port;
        const host = this.#context.host;
        const services = this.#getServices({bindPort, host});
        if (services.length === 0) {
            this.killProcess(bindPort, "SIGTERM");
            return [];
        }

        const ratholeFile = this.#writeRatholeFile({bindPort, services});
        this.ensureProcess({
            key: bindPort,
            file: "rathole",
            args: ["--server", ratholeFile],
            options: {cwd: "src/king", env: {RUST_LOG: "warn"}},
        });

        return services.map((s) => s["service_id"]);
    }

    async stateChanged() {
        let readyServices = [];
        for (const ratholeCnf of this.#context.config.ratholes) {
            readyServices = readyServices.concat(await this.#each(ratholeCnf));
        }
        this.#context.readyServiceIds = readyServices;
    }
}
