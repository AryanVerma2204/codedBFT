# plot_results.py
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import sys

def plot_scalability(df, output_file="scalability.png"):
    """Plots throughput vs. number of nodes with confidence intervals."""
    plt.style.use('seaborn-v0_8-whitegrid')
    fig, ax = plt.subplots(figsize=(10, 6))

    scalability_df = df[df['experiment_name'] == 'scalability'].copy()
    if scalability_df.empty:
        print("Warning: No scalability data found.")
        return

    scalability_df['throughput_mbps'] = scalability_df['throughput_bps'] / 1_000_000

    sns.lineplot(data=scalability_df, x='num_nodes', y='throughput_mbps', hue='protocol', style='protocol', markers=True, dashes=False, ax=ax, lw=2.5, err_style="band")

    ax.set_title('Scalability: Throughput vs. Number of Nodes', fontsize=16, fontweight='bold')
    ax.set_xlabel('Number of Nodes (n)', fontsize=12)
    ax.set_ylabel('Throughput (Mbps)', fontsize=12)
    ax.legend(title='Protocol', fontsize=11)
    ax.set_xticks(scalability_df['num_nodes'].unique())
    ax.grid(True, which='both', linestyle='--', linewidth=0.5)
    plt.tight_layout()
    plt.savefig(output_file, dpi=300)
    print(f"Saved scalability plot to {output_file}")
    plt.close()

def plot_packet_loss(df, output_file="packet_loss.png"):
    """Plots throughput vs. packet loss with confidence intervals."""
    plt.style.use('seaborn-v0_8-whitegrid')
    fig, ax = plt.subplots(figsize=(10, 6))
    
    loss_df = df[df['experiment_name'] == 'packet_loss'].copy()
    if loss_df.empty:
        print("Warning: No packet loss data found.")
        return

    loss_df['throughput_mbps'] = loss_df['throughput_bps'] / 1_000_000
    loss_df['packet_loss_pct'] = loss_df['packet_loss'] * 100

    sns.lineplot(data=loss_df, x='packet_loss_pct', y='throughput_mbps', hue='protocol', style='protocol', markers=True, dashes=False, ax=ax, lw=2.5, err_style="band")

    ax.set_title('Resilience: Throughput vs. Packet Loss', fontsize=16, fontweight='bold')
    ax.set_xlabel('Packet Loss (%)', fontsize=12)
    ax.set_ylabel('Throughput (Mbps)', fontsize=12)
    ax.legend(title='Protocol', fontsize=11)
    ax.grid(True, which='both', linestyle='--', linewidth=0.5)
    plt.tight_layout()
    plt.savefig(output_file, dpi=300)
    print(f"Saved packet loss plot to {output_file}")
    plt.close()

def plot_ablation(df, output_file="ablation.png"):
    """Creates a bar chart for the ablation study with error bars."""
    plt.style.use('seaborn-v0_8-whitegrid')
    fig, ax = plt.subplots(figsize=(8, 6))

    ablation_df = df[df['experiment_name'] == 'ablation'].copy()
    if ablation_df.empty:
        print("Warning: No ablation data found.")
        return
        
    ablation_df['throughput_mbps'] = ablation_df['throughput_bps'] / 1_000_000
    
    sns.barplot(data=ablation_df, x='protocol', y='throughput_mbps', ax=ax, palette='viridis', ci='sd', capsize=.1)

    ax.set_title('Ablation Study (at 2% Packet Loss)', fontsize=16, fontweight='bold')
    ax.set_xlabel('Protocol Variant', fontsize=12)
    ax.set_ylabel('Throughput (Mbps)', fontsize=12)
    ax.set_xticklabels(ax.get_xticklabels(), rotation=0)
    plt.tight_layout()
    plt.savefig(output_file, dpi=300)
    print(f"Saved ablation plot to {output_file}")
    plt.close()

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python plot_results.py <path_to_results.csv>")
        sys.exit(1)
        
    results_file = sys.argv[1]
    try:
        df = pd.read_csv(results_file)
    except FileNotFoundError:
        print(f"Error: Results file not found at '{results_file}'")
        sys.exit(1)

    plot_scalability(df)
    plot_packet_loss(df)
    plot_ablation(df)
