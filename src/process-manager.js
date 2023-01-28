import {execa} from "execa";

export class ProcessManager {

    #processMap = new Map();
    #serviceType;

    constructor(context, serviceType) {
        this.context = context;
        this.#serviceType = serviceType;
    }

    killProcesses(signal) {
        this.#processMap.forEach(p => p.kill(signal));
    }

    ensureProcess({key, file, args, options}) {
        if (this.#processMap.has(key)) return;

        const p = execa(file, args, options);
        console.log(`msg="Started ${p.spawnfile} ${p.spawnargs.join(" ")} service.type=${this.#serviceType}`);
        p.stdout.pipe(process.stdout);
        p.stderr.pipe(process.stderr);
        p.on("exit", async(code) => {
            console.info(`msg="Exiting ${p.spawnfile} ${p.spawnargs.join(" ")}" process.exit_code=${code} service.type=${this.#serviceType}`);
            this.#processMap.delete(key);
        });
        p.on("error", (err) => {
            console.error(`msg="Exiting ${p.spawnfile} ${p.spawnargs.join(" ")}" error.message=${err.message} service.type=${this.#serviceType}`);
            process.exit(1);
        });
        this.#processMap.set(key, p);
    }
}
