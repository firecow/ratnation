import {execa, Options} from "execa";
import waitFor from "p-wait-for";
import split2 from "split2";
import {Logger} from "./logger.js";
import {Transform} from "stream";
import {ChildProcess} from "child_process";

interface ProcessManagerEnsureProcessOpts {
    key: string;
    file: string;
    args: string[];
    options: Options;
    initTransform: () => Transform;
}

export class ProcessManager {

    public readonly logger;
    private readonly processMap = new Map<string, ChildProcess>();
    private readonly serviceType;

    constructor ({logger, serviceType}: {logger: Logger; serviceType: string}) {
        this.logger = logger;
        this.serviceType = serviceType;
    }

    processKeys () {
        return this.processMap.keys();
    }

    async killProcesses (signal: NodeJS.Signals) {
        const proms = [];
        for (const key of this.processMap.keys()) {
            proms.push(this.killProcess(key, signal));
        }
        return Promise.all(proms);
    }

    async killProcess (key: string, signal: NodeJS.Signals): Promise<void> {
        const p = this.processMap.get(key);
        if (!p) return;
        p.kill(signal);

        await waitFor(() => p.exitCode != null);
    }

    ensureProcess ({key, file, args, options, initTransform}: ProcessManagerEnsureProcessOpts) {
        if (this.processMap.has(key)) return;

        const logger = this.logger;
        const p = execa(file, args, options);
        logger.info(`Started ${p.spawnargs.join(" ")}`, {"service.type": this.serviceType});
        p.stdout?.pipe(split2()).pipe(initTransform()).pipe(process.stdout);
        p.stderr?.pipe(split2()).pipe(initTransform()).pipe(process.stderr);

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
