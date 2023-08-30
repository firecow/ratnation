type ContextVal = string | number | boolean | null | undefined;

export class Logger {

    info (msg: string, ctx: {[key: string]: ContextVal} = {}) {
        process.stdout.write(JSON.stringify({
            "@timestamp": new Date().toISOString(),
            "log.level": "info",
            "message": msg,
            ...ctx,
        }) + "\n");
    }

    error (msg: string, ctx: {[key: string]: ContextVal} = {}) {
        process.stderr.write(JSON.stringify({
            "@timestamp": new Date().toISOString(),
            "log.level": "error",
            "message": msg,
            ...ctx,
        }) + "\n");
    }
}
