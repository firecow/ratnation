import fs from "fs";
import {ProcessManager} from "../process-manager.js";

export class LingTraefikManager extends ProcessManager{

    constructor(context) {
        super(context, "ling-traefik-manager");
    }

    #findServices({name}) {
        const state = this.context.state;
        return state["services"].filter(s => {
            const ling = state["lings"].find(l => l["ling_id"] === s["ling_id"]);
            if (!ling || ling["shutting_down"]) {
                return false;
            }
            return s["name"] === name && s["king_ready"] && s["ling_ready"];
        });
    }

    #writeTraefikFile({bindPort, name, services}) {
        const lines = [
            "[tcp.routers.default]",
            "rule = \"HostSNI(`*`)\"",
            `service = "${name}"`,
            "",
        ];

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
        const services = this.#findServices({name});
        if (services.length === 0) return [];

        const traefikFile = this.#writeTraefikFile({bindPort, name, services});
        this.ensureProcess({
            key: bindPort,
            file: "traefik",
            args: [`--entrypoints.tcp.address=:${bindPort}/tcp`, `--providers.file.filename=${traefikFile}`, "--providers.file.watch=true", "--log.level=error"],
            options: {cwd: "src/ling"}
        });
    }

    async stateChanged() {
        for (const proxyCnf of this.context.config.proxyMap.values()) {
            this.#each(proxyCnf);
        }
    }
}
