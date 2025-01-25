import {
  HeadingXSmall,
  PRIMITIVE_COLORS,
  ParagraphSmall,
  StyledTableSemantic,
  StyledTableSemanticBody,
  StyledTableSemanticBodyCell,
  StyledTableSemanticBodyRow,
  StyledTableSemanticHead,
  StyledTableSemanticHeadCell,
  StyledTableSemanticHeadRow,
  StyledTableSemanticRoot,
} from "@nilfoundation/ui-kit";
import { useUnit } from "effector-react";
import { useStyletron } from "styletron-react";
import { blockRoute } from "../../../routing/routes/blockRoute";
import { StyledList } from "../../../shared";
import { InfoContainer } from "../../../shared/components/InfoContainer";
import { Link } from "../../../shared/components/Link";
import { useMobile } from "../../../shared/hooks/useMobile";
import { $latestBlocks, fetchLatestBlocksFx } from "../../models/model";
import { styles as s } from "./styles";

export const Blocks = () => {
  const [blockslist] = useUnit([$latestBlocks, fetchLatestBlocksFx.pending]);
  const [css] = useStyletron();
  const [isMobile] = useMobile();

  const fontStyle = css({
    fontSize: "12px!important",
  });

  return (
    <InfoContainer title="Latest Blocks">
      {!isMobile && (
        <div className={css(s.table)}>
          <StyledTableSemanticRoot>
            <StyledTableSemantic
              data-testid="blocks-table"
              style={{
                tableLayout: "fixed",
                width: "100%",
              }}
            >
              <StyledTableSemanticHead>
                <StyledTableSemanticHeadRow>
                  <StyledTableSemanticHeadCell style={{ width: "10%" }}>
                    Shard
                  </StyledTableSemanticHeadCell>
                  <StyledTableSemanticHeadCell style={{ width: "10%" }}>
                    Height
                  </StyledTableSemanticHeadCell>
                  <StyledTableSemanticHeadCell style={{ width: "10%" }}>
                    Txn Count
                  </StyledTableSemanticHeadCell>
                </StyledTableSemanticHeadRow>
              </StyledTableSemanticHead>
              <StyledTableSemanticBody>
                {blockslist.map(({ id, hash, shard_id, in_txn_num }) => {
                  return (
                    <StyledTableSemanticBodyRow key={hash}>
                      <StyledTableSemanticBodyCell>{shard_id}</StyledTableSemanticBodyCell>
                      <StyledTableSemanticBodyCell>
                        <Link to={blockRoute} params={{ shard: shard_id, id }}>
                          {id}
                        </Link>
                      </StyledTableSemanticBodyCell>
                      <StyledTableSemanticBodyCell>
                        <Link to={blockRoute} params={{ shard: shard_id, id }}>
                          {in_txn_num}
                        </Link>
                      </StyledTableSemanticBodyCell>
                    </StyledTableSemanticBodyRow>
                  );
                })}
              </StyledTableSemanticBody>
            </StyledTableSemantic>
          </StyledTableSemanticRoot>
        </div>
      )}
      {isMobile && (
        <StyledList scrollable>
          {blockslist.map(({ id, hash, shard_id, in_txn_num }) => {
            return (
              <StyledList.Item key={hash}>
                <HeadingXSmall color={PRIMITIVE_COLORS.gray400} className={fontStyle}>
                  Shard
                </HeadingXSmall>
                <ParagraphSmall color={PRIMITIVE_COLORS.gray200}>{shard_id}</ParagraphSmall>
                <HeadingXSmall color={PRIMITIVE_COLORS.gray400} className={fontStyle}>
                  Height
                </HeadingXSmall>
                <ParagraphSmall color={PRIMITIVE_COLORS.gray200}>
                  <Link to={blockRoute} params={{ shard: shard_id, id }}>
                    {id}
                  </Link>
                </ParagraphSmall>
                <HeadingXSmall color={PRIMITIVE_COLORS.gray400} className={fontStyle}>
                  Txs count
                </HeadingXSmall>
                <ParagraphSmall color={PRIMITIVE_COLORS.gray200}>
                  <Link to={blockRoute} params={{ shard: shard_id, id }}>
                    {in_txn_num}
                  </Link>
                </ParagraphSmall>
              </StyledList.Item>
            );
          })}
        </StyledList>
      )}
    </InfoContainer>
  );
};
