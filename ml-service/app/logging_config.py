import logging
import os
import sys


class ServiceContextFilter(logging.Filter):
    def filter(self, record: logging.LogRecord) -> bool:
        record.service = "ml-service"
        return True


def configure_logging() -> logging.Logger:
    level_name = os.getenv("LOG_LEVEL", "INFO").upper()
    level = getattr(logging, level_name, logging.INFO)
    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(
        logging.Formatter(
            '{"timestamp":"%(asctime)s","level":"%(levelname)s",'
            '"service":"%(service)s","event":"%(message)s"}'
        )
    )
    handler.addFilter(ServiceContextFilter())
    logger = logging.getLogger("meloop.ml")
    logger.setLevel(level)
    logger.handlers.clear()
    logger.addHandler(handler)
    logger.propagate = False
    return logger
