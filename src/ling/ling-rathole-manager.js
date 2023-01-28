import {execa} from "execa";
import fs from "fs";

export class LingRatholeManager {
    ratholeProcessMap = new Map();

    constructor(context) {
        this.context = context;
    }

    killProcesses(signal) {
        this.ratholeProcessMap.forEach(rathole => {
            rathole.kill(signal);
        });
    }

    #ensureRathole({kingBindAddr, ratholeFile}) {
        if (this.ratholeProcessMap.has(kingBindAddr)) return;

        const rathole = execa("rathole", ["--client", ratholeFile], {cwd: "src/ling/", env: {RUST_LOG: "warn"}});
        console.log(`msg="Started rathole client" service.type=ratling log.logger=rathole-manager remote_addr=${kingBindAddr} pid=${rathole.pid}`);
        rathole.stdout.pipe(process.stdout);
        rathole.stderr.pipe(process.stderr);
        rathole.on("exit", async(code) => {
            console.info(`msg="Rathole exited" process_exit_code=${code} service.type=ratling log.logger=rathole-manager`);
            this.ratholeProcessMap.delete(kingBindAddr);
        });

        this.ratholeProcessMap.set(kingBindAddr, rathole);
    }

    async stateChanged() {
        const state = this.context.state;
        const config = this.context.config;
        const lingId = this.context.lingId;
        const services = state["services"].filter(s => config.ratholeMap.has(s["name"]) && s["ling_id"] === lingId && s["king_ready"] === true);
        for (const kingBindAddr of services.map(s => `${s["host"]}:${s["bind_port"]}`)) {
            const lines = [
                "[client]",
                `remote_addr = "${kingBindAddr}"`,
                "",
                "[client.services]",
            ];

            for (const service of services.filter(s => `${s["host"]}:${s["bind_port"]}` === kingBindAddr)) {
                const ratholeCnf = config.ratholeMap.get(service["name"]);
                lines.push(`[client.services.${service["service_id"].replace(/:/g, "-")}]`);
                lines.push(`token = "${service["token"]}"`);
                lines.push(`local_addr = "${ratholeCnf["local_addr"]}"`);
            }

            const ratholeFile = `rathole-client-${lingId.replace(/:/g, "-")}.toml`;
            fs.writeFileSync(`src/ling/${ratholeFile}`, `${lines.join("\n")}\n`, "utf8");
            this.#ensureRathole({kingBindAddr, ratholeFile});
        }
        this.context.readyServiceIds = services.map(s => s["service_id"]);
    }
}
