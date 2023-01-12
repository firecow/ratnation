import fs from "fs";
import {execa} from "execa";

export class KingRatholeManager {
    ratholeProcessMap = new Map();

    constructor(context) {
        this.context = context;
    }

    #ensureRathole({bindPort, ratholeFile}) {
        if (this.ratholeProcessMap.has(bindPort)) return;

        const rathole = execa("rathole", [ratholeFile], {cwd: "src/king/"});
        console.log(`msg="Started rathole" service_type=ratking bind_addr=${bindPort} pid=${rathole.pid}`);
        rathole.stdout.pipe(process.stdout);
        rathole.stderr.pipe(process.stderr);
        rathole.on("exit", (code) => {
            console.info(`msg="Rathole exited" process_exit_code=${code} service_type=ratking log.logger=rathole-manager`);
            this.ratholeProcessMap.delete(bindPort);
            this.doit();
        });

        this.ratholeProcessMap.set(bindPort, rathole);
    }

    async #each(ratholeCnf) {
        const state = this.context.state;
        const bindPort = ratholeCnf.bind_port;
        const services = state.services.filter(s => s["bind_port"] === ratholeCnf.bind_port);
        if (services.length === 0) return [];

        const lines = [
            "[server]",
            `bind_addr = "0.0.0.0:${bindPort}"`,
            "",
            "[server.services]",
        ];

        for (const service of services) {
            lines.push(`[server.services.${service["name"]}]`);
            lines.push(`token = "${service["token"]}"`);
            lines.push(`bind_addr = "0.0.0.0:${service["remote_port"]}"`);
        }

        const ratholeFile = `server-${bindPort}.toml`;
        await fs.promises.writeFile(`src/king/${ratholeFile}`, `${lines.join("\n")}\n`, "utf8");
        this.#ensureRathole({bindPort, ratholeFile});

        return services.map((s) => s["name"]);
    }

    async doit() {
        let readyServices = [];
        for (const ratholeCnf of this.context.config.ratholes) {
            readyServices = readyServices.concat(await this.#each(ratholeCnf));
        }
        this.context.readyServices = readyServices;
    }
}
