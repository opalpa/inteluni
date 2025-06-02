#!/usr/bin/env python
import glob
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import numpy as np

# Find all CSV files
all_files = glob.glob("run_*.csv")
if not all_files:
    raise RuntimeError("No files matched 'run_*.csv'. Are you in the right directory?")

print(f"Found {len(all_files)} files.")

# Read and combine CSVs
df = pd.concat((pd.read_csv(f) for f in all_files), ignore_index=True)

# Example plot: TauL by foresight and complexity
pivot_tau = df.pivot_table(
    values="TauL",
    index="foresight",
    columns="complexity",
    aggfunc="mean"
)

plt.figure(figsize=(10, 6))
sns.heatmap(pivot_tau, annot=True, fmt=".1f", cmap="viridis")
plt.title("Average TauL by Foresight and Complexity")
plt.xlabel("Complexity")
plt.ylabel("Foresight")
plt.tight_layout()
plt.savefig("tau_heatmap.png")
plt.show()

# Add deltaC metric if not already there
df["deltaC"] = (df["C_react"] - df["C_pred"]) / df["C_react"].clip(lower=1)

# Heatmap of deltaC: foresight vs complexity
pivot_dc = df.pivot_table(
    values="deltaC",
    index="foresight",
    columns="complexity",
    aggfunc="mean"
)

plt.figure(figsize=(10, 6))
sns.heatmap(pivot_dc, annot=True, fmt=".2f", cmap="magma")
plt.title("Forecast Advantage (ΔC) by Foresight and Complexity")
plt.xlabel("Complexity")
plt.ylabel("Foresight")
plt.tight_layout()
plt.savefig("deltaC_heatmap.png")
plt.show()



df["deltaC"] = (df["C_react"] - df["C_pred"]) / df["C_react"].clip(lower=1)
df["logTauL"] = np.log10(df["TauL"].replace(0, np.nan))

plt.figure(figsize=(8, 5))
sns.lineplot(data=df, x="noise", y="deltaC", errorbar="sd")
plt.title("Forecast Advantage vs Noise")
plt.xlabel("Noise")
plt.ylabel("Forecast Advantage (ΔC)")
plt.grid(True)
plt.tight_layout()
plt.savefig("deltaC_vs_noise.png")
plt.show()


plt.figure(figsize=(8, 6))
sc = plt.scatter(df["K"], df["deltaC"], c=df["TauL"], cmap="plasma", alpha=0.7)
plt.colorbar(sc, label="TauL")
plt.xlabel("Kolmogorov Proxy (K)")
plt.ylabel("Forecast Advantage (ΔC)")
plt.title("ΔC vs K colored by TauL")
plt.grid(True)
plt.tight_layout()
plt.savefig("deltaC_vs_K_TauL.png")
plt.show()


plt.figure(figsize=(8, 5))
sns.lineplot(data=df, x="noise", y="TauL", errorbar="sd")
plt.title("Lyapunov Horizon vs Noise")
plt.xlabel("Noise")
plt.ylabel("TauL")
plt.grid(True)
plt.tight_layout()
plt.savefig("tauL_vs_noise.png")
plt.show()


df_ridge = df[(df["TauL"] > 3) & (df["TauL"] < 15)]

plt.figure(figsize=(8, 5))
sns.boxplot(data=df_ridge, x="foresight", y="deltaC")
plt.title("ΔC vs Foresight in Forecast-Useful Zone")
plt.xlabel("Foresight Depth")
plt.ylabel("Forecast Advantage (ΔC)")
plt.grid(True)
plt.tight_layout()
plt.savefig("deltaC_vs_foresight_ridge.png")
plt.show()




