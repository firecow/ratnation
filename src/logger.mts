export class Logger {

    info (msg: string, ctx: {[key: string]: string | number | boolean | null | undefined} = {}) {
        console.info(JSON.stringify({
            "@timestamp": new Date().toISOString(),
            "log.level": "info",
            "message": msg,
            ...ctx,
        }));
    }

    error (msg: string, ctx: {[key: string]: string | number | boolean | null | undefined} = {}) {
        console.error(JSON.stringify({
            "@timestamp": new Date().toISOString(),
            "log.level": "error",
            "message": msg,
            ...ctx,
        }));
    }
}
