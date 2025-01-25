import { LabelMedium } from "baseui/typography";
import { useStyletron } from "styletron-react";
import { formatNumber } from "../../../shared";
import { type SHARD_WORKLOAD, getBackgroundBasedOnWorkload } from "../../types/SHARD_WORKLOAD";
import { styles } from "./styles";

type ShardProps = {
  workload: SHARD_WORKLOAD;
  txCount: number;
};

export const Shard = ({ workload, txCount }: ShardProps) => {
  const [css] = useStyletron();

  return (
    <div
      className={css({ ...styles.shard, backgroundColor: getBackgroundBasedOnWorkload(workload) })}
    >
      <LabelMedium>{formatNumber(txCount)}</LabelMedium>
    </div>
  );
};
