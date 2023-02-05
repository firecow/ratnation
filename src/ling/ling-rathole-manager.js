import fs from "fs";
import {ProcessManager} from "../process-manager.js";

export class LingRatholeManager extends ProcessManager {

    #context;

    constructor(context) {
        super("ling-rathole-manager");
        this.#context = context;
    }

    #getServices({config, lingId}) {
        const state = this.#context.state;
        return state["services"].filter(s => {
            const king = state["kings"].find(k => k["host"] === s["host"] && k["bind_port"] === s["bind_port"]);
            if (king && king["shutting_down"]) {
                return false;
            }
            return config.ratholeMap.has(s["name"]) && s["ling_id"] === lingId && s["king_ready"] === true;
        });
    }

    #writeRatholeFile({kingBindAddr, services, config, lingId}) {
        const lines = [];
        lines.push(
            "[client]",
            `remote_addr = "${kingBindAddr}"`,
            "",
            "[client.services]",
        );

        for (const service of services.filter(s => `${s["host"]}:${s["bind_port"]}` === kingBindAddr)) {
            const ratholeCnf = config.ratholeMap.get(service["name"]);
            lines.push(
                `[client.services.${service["service_id"].replace(/:/g, "-")}]`,
                `token = "${service["token"]}"`,
                `local_addr = "${ratholeCnf["local_addr"]}"`
            );
        }

        const ratholeFile = `rathole-client-${lingId.replace(/:/g, "-")}.toml`;
        fs.writeFileSync(`src/ling/${ratholeFile}`, `${lines.join("\n")}\n`, "utf8");
        return ratholeFile;
    }

    async stateChanged() {
        const config = this.#context.config;
        const lingId = this.#context.lingId;
        const services = this.#getServices({config, lingId});

        const kingBindAddrs = services.map(s => `${s["host"]}:${s["bind_port"]}`);

        // Ensure rathole process is running and maintaing rathole client configuration file
        for (const kingBindAddr of kingBindAddrs) {
            const ratholeFile = this.#writeRatholeFile({kingBindAddr, services, config, lingId});
            this.ensureProcess({
                key: kingBindAddr,
                file: "rathole",
                args: ["--client", ratholeFile],
                options: {cwd: "src/ling/", env: {RUST_LOG: "warn"}}
            });
        }

        // Kill processes if king bind_addr doesn't have any ports active.
        for (const kingBindAddr of this.processKeys()) {
            if (kingBindAddrs.includes(kingBindAddr)) continue;
            this.killProcess(kingBindAddr, "SIGTERM");
        }

        this.#context.readyServiceIds = services.map(s => s["service_id"]);
    }
}
