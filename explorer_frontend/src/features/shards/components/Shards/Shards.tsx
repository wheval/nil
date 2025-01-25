import { COLORS, LabelSmall, ParagraphLarge, SPACE, Spinner } from "@nilfoundation/ui-kit";
import { useUnit } from "effector-react";
import { useStyletron } from "styletron-react";
import { formatNumber } from "../../../shared";
import { InfoContainer } from "../../../shared/components/InfoContainer";
import { $shards, fetchShardsFx } from "../../models/model";
import { SHARD_WORKLOAD } from "../../types/SHARD_WORKLOAD";
import { Shard } from "./Shard";
import { styles as s } from "./styles";

export const Shards = () => {
  const [shards, pending] = useUnit([$shards, fetchShardsFx.pending]);
  const [css] = useStyletron();

  return (
    <InfoContainer title="Shards">
      <LabelSmall paddingTop={SPACE[24]} color={COLORS.gray300}>
        Transactions
      </LabelSmall>
      <ParagraphLarge paddingTop={SPACE[8]}>
        {formatNumber(shards.reduce((acc, { tx_count }) => acc + tx_count, 0))}
      </ParagraphLarge>
      {pending && !shards.length ? (
        <Spinner />
      ) : (
        <div className={css(s.shardsContainer)} data-testid="shards-container">
          {shards.map(({ shard_id, tx_count }) => (
            <Shard key={shard_id} txCount={tx_count} workload={getShardWorkload(tx_count)} />
          ))}
        </div>
      )}
    </InfoContainer>
  );
};

const getShardWorkload = (txCount: number): SHARD_WORKLOAD => {
  if (txCount < 100) {
    return SHARD_WORKLOAD.low;
    // biome-ignore lint/style/noUselessElse: <explanation>
  } else if (txCount < 500) {
    return SHARD_WORKLOAD.medium;
  }

  return SHARD_WORKLOAD.high;
};
