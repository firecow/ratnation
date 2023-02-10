import {execa, ExecaChildProcess, Options} from "execa";
import {Logger} from "./logger.mjs";

interface ProcessManagerEnsureProcessOpts {
    key: string;
    file: string;
    args: string[];
    options: Options;
}

export class ProcessManager {

    public readonly logger;
    private readonly processMap = new Map<string, ExecaChildProcess>();
    private readonly serviceType;

    constructor ({logger, serviceType}: {logger: Logger; serviceType: string}) {
        this.logger = logger;
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

        const logger = this.logger;
        const p = execa(file, args, options);
        logger.info(`Started ${p.spawnargs.join(" ")}`, {"service.type": this.serviceType});
        p.stdout?.pipe(process.stdout);
        p.stderr?.pipe(process.stderr);

        void p.once("exit", (code) => {
            logger.info(`Exiting ${p.spawnargs.join(" ")}`, {"service.type": this.serviceType, "process.exit_code": code});
            this.processMap.delete(key);
            void p.removeAllListeners();
        });

        void p.once("error", (err) => {
            logger.info(`Error ${p.spawnargs.join(" ")}`, {"service.type": this.serviceType, "error.message": err.message});
            process.exit(1);
        });

        this.processMap.set(key, p);
    }
}
