import {execa} from "execa";

export class ProcessManager {

    #processMap = new Map();
    #serviceType;

    constructor(serviceType) {
        this.#serviceType = serviceType;
    }

    processKeys() {
        return this.#processMap.keys();
    }

    killProcesses(signal) {
        this.#processMap.forEach(p => {
            p.stdout.unpipe(process.stdout);
            p.stderr.unpipe(process.stderr);
            p.kill(signal);
        });
    }

    killProcess(key, signal) {
        const p = this.#processMap.get(key);
        if (!p) return false;
        p.stdout.unpipe(process.stdout);
        p.stderr.unpipe(process.stderr);
        p.kill(signal);
    }

    ensureProcess({key, file, args, options}) {
        if (this.#processMap.has(key)) return;

        const p = execa(file, args, options);
        console.log(`message="Started ${p.spawnargs.join(" ")}" service.type=${this.#serviceType}`);
        p.stdout.pipe(process.stdout);
        p.stderr.pipe(process.stderr);
        p.on("exit", async(code) => {
            console.info(`message="Exiting ${p.spawnargs.join(" ")}" process.exit_code=${code} service.type=${this.#serviceType}`);
            this.#processMap.delete(key);
        });
        p.on("error", (err) => {
            console.error(`message="Exiting ${p.spawnargs.join(" ")}" error.message=${err.message} service.type=${this.#serviceType}`);
            process.exit(1);
        });
        this.#processMap.set(key, p);
    }
}
