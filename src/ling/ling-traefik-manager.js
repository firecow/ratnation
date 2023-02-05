import fs from "fs";
import {ProcessManager} from "../process-manager.js";

export class LingTraefikManager extends ProcessManager {

    #context;

    constructor(context) {
        super("ling-traefik-manager");
        this.#context = context;
    }

    #getServices({name}) {
        const state = this.#context.state;
        return state["services"].filter(s => {
            if (s["name"] !== name) {
                return false;
            }

            const ling = state["lings"].find(l => l["ling_id"] === s["ling_id"]);
            if (!ling || ling["shutting_down"] || !s["ling_ready"]) {
                return false;
            }

            const king = state["kings"].find(k => k["host"] === s["host"] && k["bind_port"] === s["bind_port"]);
            if (!king || king["shutting_down"] || !s["king_ready"]) {
                return false;
            }

            return true;
        });
    }

    #writeTraefikFile({bindPort, name, services}) {
        const lines = [];
        lines.push(
            "[tcp.routers.default]",
            "rule = \"HostSNI(`*`)\"",
            `service = "${name}"`,
            "",
        );

        for (const service of services) {
            lines.push(
                `[[tcp.services.${name}.loadBalancer.servers]]`,
                `address = "${service["host"]}:${service["remote_port"]}"`,
            );
        }

        const traefikFile = `traefik-${bindPort}.toml`;
        fs.writeFileSync(`src/ling/${traefikFile}`, `${lines.join("\n")}\n`, "utf8");
        return traefikFile;
    }

    #each(proxyCnf) {
        const bindPort = proxyCnf["bind_port"];
        const name = proxyCnf["name"];
        const services = this.#getServices({name});
        if (services.length === 0) {
            return this.killProcess(bindPort, "SIGTERM");
        }

        const traefikFile = this.#writeTraefikFile({bindPort, name, services});
        this.ensureProcess({
            key: bindPort,
            file: "traefik",
            args: [`--entrypoints.tcp.address=:${bindPort}/tcp`, `--providers.file.filename=${traefikFile}`, "--providers.file.watch=true", "--log.level=error"],
            options: {cwd: "src/ling"},
        });
    }

    async stateChanged() {
        for (const proxyCnf of this.#context.config.proxyMap.values()) {
            this.#each(proxyCnf);
        }
    }
}
