import fs from "fs";
import {ProcessManager} from "../process-manager.js";
import {StateService} from "../state-handler.js";
import {LingProxyConfig} from "./ling-config.js";
import {LingContext} from "./ling-cmd.js";
import {TraefikTransform} from "../stream/traefik-transform.js";

export class LingTraefikManager extends ProcessManager {

    private readonly context;

    constructor (context: LingContext) {
        super({...context, serviceType: "ratling"});
        this.context = context;
    }

    private getServices ({name}: {name: string}) {
        const state = this.context.state;
        return state.services.filter((s) => {
            if (s.name !== name) {
                return false;
            }

            const ling = state.lings.find((l) => l.ling_id === s.ling_id);
            if (!ling || ling.shutting_down || !s.ling_ready) {
                return false;
            }

            const king = state.kings.find((k) => k.host === s.host && k.bind_port === s.bind_port);
            return (king && !king.shutting_down && s.king_ready);
        });
    }

    private writeTraefikFile ({bindPort, name, services}: {bindPort: number; name: string; services: StateService[]}) {
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
                `address = "${service.host}:${service.remote_port}"`,
            );
        }

        const traefikFile = `traefik-${bindPort}.toml`;
        fs.writeFileSync(`src/ling/${traefikFile}`, `${lines.join("\n")}\n`, "utf8");
        return traefikFile;
    }

    async each (proxyCnf: LingProxyConfig) {
        const bindPort = proxyCnf["bind_port"];
        const name = proxyCnf["name"];
        const services = this.getServices({name});
        if (services.length === 0) {
            return this.killProcess(`${bindPort}`, "SIGTERM");
        }

        const traefikFile = this.writeTraefikFile({bindPort, name, services});
        this.ensureProcess({
            key: `${bindPort}`,
            file: "traefik",
            args: [`--entrypoints.tcp.address=:${bindPort}/tcp`, `--providers.file.filename=${traefikFile}`, "--providers.file.watch=true", "--log.level=error"],
            options: {cwd: "src/ling"},
            initTransform () {
                return new TraefikTransform();
            },
        });
    }

    async stateChanged () {
        const proms = [];
        for (const proxyCnf of this.context.config.proxyMap.values()) {
            proms.push(this.each(proxyCnf));
        }
        await Promise.all(proms);
    }
}
