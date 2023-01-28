import fs from "fs";
import {execa} from "execa";

export class KingRatholeManager {
    ratholeProcessMap = new Map();

    constructor(context) {
        this.context = context;
    }

    #ensureRathole({bindPort, ratholeFile}) {
        if (this.ratholeProcessMap.has(bindPort)) return;

        const rathole = execa("rathole", [ratholeFile], {cwd: "src/king/", env: {RUST_LOG: "warn"}});
        console.log(`msg="Started rathole server" service.type=ratking bind_addr=${bindPort} pid=${rathole.pid}`);
        rathole.stdout.pipe(process.stdout);
        rathole.stderr.pipe(process.stderr);
        rathole.on("exit", async(code) => {
            console.info(`msg="Rathole exited" process_exit_code=${code} service.type=ratking log.logger=rathole-manager`);
            this.ratholeProcessMap.delete(bindPort);
            await this.stateChanged();
        });

        this.ratholeProcessMap.set(bindPort, rathole);
    }

    async #each(ratholeCnf) {
        const state = this.context.state;
        const bindPort = ratholeCnf.bind_port;
        const services = state.services.filter(s => {
            if (s["bind_port"] !== ratholeCnf.bind_port) {
                return false;
            }

            const ling = state.lings.find(l => l["ling_id"] === s["ling_id"]);
            if (!ling) {
                console.error(`msg="Ling not found for service" service_name=${s["name"]} service.type=ratking log.logger=rathole-manager`);
                return false;
            }

            return !ling["shutting_down"];
        });
        if (services.length === 0) return [];

        const lines = [
            "[server]",
            `bind_addr = "0.0.0.0:${bindPort}"`,
            "",
            "[server.services]",
        ];

        for (const service of services) {
            lines.push(`[server.services.${service["service_id"].replace(/:/g, "-")}]`);
            lines.push(`token = "${service["token"]}"`);
            lines.push(`bind_addr = "0.0.0.0:${service["remote_port"]}"`);
        }

        const ratholeFile = `rathole-server-${bindPort}.toml`;
        fs.writeFileSync(`src/king/${ratholeFile}`, `${lines.join("\n")}\n`, "utf8");
        this.#ensureRathole({bindPort, ratholeFile});

        return services.map((s) => s["service_id"]);
    }

    async stateChanged() {
        let readyServices = [];
        for (const ratholeCnf of this.context.config.ratholes) {
            readyServices = readyServices.concat(await this.#each(ratholeCnf));
        }
        this.context.readyServiceIds = readyServices;
    }
}
