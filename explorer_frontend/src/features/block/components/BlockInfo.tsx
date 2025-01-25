import { HeadingXLarge, ParagraphMedium, SPACE, Skeleton } from "@nilfoundation/ui-kit";
import { useStyletron } from "baseui";
import { useUnit } from "effector-react";
import { Divider } from "../../shared";
import { Info } from "../../shared/components/Info";
import { InfoBlock } from "../../shared/components/InfoBlock";
import { $block, loadBlockFx } from "../models/model";

export const BlockInfo = ({
  className,
}: {
  className?: string;
}) => {
  const [blockInfo, isPending] = useUnit([$block, loadBlockFx.pending]);
  const [css] = useStyletron();

  if (isPending) {
    return (
      <div className={className}>
        <HeadingXLarge className={css({ marginBottom: SPACE[32] })}>Block</HeadingXLarge>
        <Skeleton animation rows={6} width={"300px"} height={"400px"} />
      </div>
    );
  }

  if (blockInfo) {
    return (
      <div className={className}>
        <InfoBlock>
          <Info label="Shard id" value={blockInfo.shard_id.toString()} />
          <Info label="Height" value={blockInfo.id} />
          <Info label="Hash" value={`0x${blockInfo.hash.toLowerCase()}`} />
          <Divider />
          <Info label="Incoming transactions" value={blockInfo.in_txn_num} />
          <Info label="Outgoing transactions" value={blockInfo.out_txn_num} />
        </InfoBlock>
      </div>
    );
  }

  return (
    <div className={className}>
      <HeadingXLarge>Block</HeadingXLarge>
      <InfoBlock>
        <ParagraphMedium>Block not found</ParagraphMedium>
      </InfoBlock>
    </div>
  );
};
