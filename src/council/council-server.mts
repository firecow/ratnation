import {AssertionError} from "assert";
import findmyway from "find-my-way";
import http, {IncomingMessage, ServerResponse} from "http";
import {Logger} from "../logger.mjs";
import {State} from "../state-handler.mjs";
import {CouncilStateCleaner} from "./coucil-state-cleaner.mjs";
import {CouncilProvisioner} from "./council-provisioner.mjs";
import getState from "./get-state.mjs";
import putKing from "./put-king.mjs";
import putLing from "./put-ling.mjs";
import {to} from "../utils.mjs";

export interface RouteRes { end: (str: string) => void; setHeader: (key: string, val: string) => void}
export interface RouteCtx {
    logger: Logger;
    req: IncomingMessage;
    res: RouteRes;
    provisioner: CouncilProvisioner;
    state: State;
}
type RouteFunc = (opts: RouteCtx) => Promise<void>;

export default function createServer () {
    const logger = new Logger();
    const state = {
        revision: 0,
        services: [],
        kings: [],
        lings: [],
    };

    const provisioner = new CouncilProvisioner({logger, state});
    const cleaner = new CouncilStateCleaner({logger, state});

    function initRoute (routeFunc: RouteFunc) {
        return async (req: IncomingMessage, res: ServerResponse) => {
            const [err] = await to(routeFunc({logger, req, res, state, provisioner}));
            if (err instanceof AssertionError) {
                res.setHeader("Content-Type", "text/plain; charset=utf-8");
                res.statusCode = 400;
                return res.end(err.message);
            } else if (err) {
                res.setHeader("Content-Type", "text/plain; charset=utf-8");
                res.statusCode = 500;
                return res.end(err.stack);
            }
        };
    }

    const router = findmyway({
        defaultRoute: (req, res) => {
            res.statusCode = 404;
            res.setHeader("Content-Type", "text/plain; charset=utf-8");
            res.end("Page could not be found");
        },
    });

    router.on("GET", "/state", initRoute(getState));
    router.on("PUT", "/ling", initRoute(putLing));
    router.on("PUT", "/king", initRoute(putKing));

    return {server: http.createServer((req, res) => router.lookup(req, res)), cleaner};
}
