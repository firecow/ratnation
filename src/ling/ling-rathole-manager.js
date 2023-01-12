import {execa} from "execa";
import fs from "fs";

export class LingRatholeManager {
    ratholeProcessMap = new Map();

    constructor(context) {
        this.context = context;
    }

    #ensureRathole({kingBindAddr, ratholeFile}) {
        if (this.ratholeProcessMap.has(kingBindAddr)) return;

        const rathole = execa("rathole", [ratholeFile], {cwd: "src/ling/"});
        console.log(`msg"Started rathole" service_type=ratling log.logger=socat-manager remote_addr=${kingBindAddr} pid=${rathole.pid}`);
        rathole.stdout.pipe(process.stdout);
        rathole.stderr.pipe(process.stderr);
        rathole.on("exit", (code) => {
            console.info(`msg="Rathole exited" process_exit_code=${code} service_type=ratling log.logger=rathole-manager`);
            this.ratholeProcessMap.delete(kingBindAddr);
            this.doit();
        });

        this.ratholeProcessMap.set(kingBindAddr, rathole);
    }

    async doit() {
        const state = this.context.state;
        const config = this.context.config;
        const services = state["services"].filter(s => config.ratholeMap.has(s["name"]) && s["king_ready"] === true);
        for (const kingBindAddr of services.map(s => `${s["host"]}:${s["bind_port"]}`)) {
            const lines = [
                "[client]",
                `remote_addr = "${kingBindAddr}"`,
                "",
                "[client.services]",
            ];

            for (const service of services.filter(s => `${s["host"]}:${s["bind_port"]}` === kingBindAddr)) {
                const ratholeCnf = config.ratholeMap.get(service["name"]);
                lines.push(`[client.services.${service["name"]}]`);
                lines.push(`token = "${service["token"]}"`);
                lines.push(`local_addr = "${ratholeCnf["local_addr"]}"`);
            }

            const ratholeFile = `client-${kingBindAddr.replace(/:/g, "-")}.toml`;
            await fs.promises.writeFile(`src/ling/${ratholeFile}`, `${lines.join("\n")}\n`, "utf8");

            this.#ensureRathole({kingBindAddr, ratholeFile});

        }
        this.context.readyServices = services.map(s => s["name"]);
    }
}
