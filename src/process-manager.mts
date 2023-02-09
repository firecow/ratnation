import {execa, ExecaChildProcess, Options} from "execa";

interface ProcessManagerEnsureProcessOpts {
    key: string;
    file: string;
    args: string[];
    options: Options;
}

export class ProcessManager {

    private readonly processMap = new Map<string, ExecaChildProcess>();
    private readonly serviceType;

    constructor (serviceType: string) {
        this.serviceType = serviceType;
    }

    processKeys () {
        return this.processMap.keys();
    }

    killProcesses (signal: NodeJS.Signals) {
        this.processMap.forEach(p => {
            p.stdout?.unpipe(process.stdout);
            p.stderr?.unpipe(process.stderr);
            p.kill(signal);
        });
    }

    killProcess (key: string, signal: NodeJS.Signals) {
        const p = this.processMap.get(key);
        if (!p) return false;
        p.stdout?.unpipe(process.stdout);
        p.stderr?.unpipe(process.stderr);
        p.kill(signal);
    }

    ensureProcess ({key, file, args, options}: ProcessManagerEnsureProcessOpts) {
        if (this.processMap.has(key)) return;

        const p = execa(file, args, options);
        console.log(`message="Started ${p.spawnargs.join(" ")}" service.type=${this.serviceType}`);
        p.stdout?.pipe(process.stdout);
        p.stderr?.pipe(process.stderr);

        void p.once("exit", (code) => {
            console.info(`message="Exiting ${p.spawnargs.join(" ")}" process.exit_code=${code} service.type=${this.serviceType}`);
            this.processMap.delete(key);
            void p.removeAllListeners();
        });

        void p.once("error", (err) => {
            console.error(`message="Exiting ${p.spawnargs.join(" ")}" error.message=${err.message} service.type=${this.serviceType}`);
            process.exit(1);
        });

        this.processMap.set(key, p);
    }
}
