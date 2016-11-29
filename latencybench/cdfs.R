#!/usr/bin/env Rscript

library(scales)
library(ggplot2)
library(RColorBrewer)

args <- commandArgs(trailingOnly=TRUE)

output_pdf <- args[1]
xaxis.label <- args[2]

data <- data.frame(Kind=c(), Value=c())
for (p in args[c(-1, -2)]) {
	subdata <- read.csv(p, header=FALSE)
	colnames(subdata) <- c("Value")
	subdata$Kind <- rep.int(basename(p), length(subdata$Value))
	data <- rbind(data, subdata)
}

pdf(output_pdf, height=5, width=8)

# Below gives the warning
# Removed XXXX rows containing non-finite values (stat_ecdf).
# because we bound the data and ggplot produces infinite values
# outside of that range. Since this is safe, suppress it so we
# don't worry about it.
options(warn = -1)
	ggplot(data, aes(x=Value, color=Kind)) +
		stat_ecdf() +
		xlim(NA, 8000) +
		xlab(xaxis.label) +
		ylab("Percentile") +
		scale_color_brewer(palette="Dark2")
options(warn = 0)

junk <- dev.off()
