// BKL-655: Página de Termos de Uso StoreMake (CDC Art. 14 marketplace)

export function TermosUsoPage() {
  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-3xl mx-auto bg-white rounded-lg shadow p-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">Termos de Uso — StoreMake</h1>
        <p className="text-sm text-gray-500 mb-8">Versão 1.0 · Vigência: 10 de abril de 2026</p>

        <section className="mb-8">
          <h2 className="text-xl font-semibold text-gray-800 mb-3">1. Aceitação dos Termos</h2>
          <p className="text-gray-600 leading-relaxed">
            Ao criar uma loja ou realizar uma compra no StoreMake, você declara que é maior de 18 anos,
            leu e concorda com estes Termos de Uso e tem capacidade legal para celebrar contratos.
            O uso continuado após alterações implica aceitação dos novos termos.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-xl font-semibold text-gray-800 mb-3">2. Partes e Papéis</h2>
          <div className="overflow-x-auto">
            <table className="w-full border-collapse border border-gray-200 text-sm">
              <thead>
                <tr className="bg-gray-50">
                  <th className="border border-gray-200 px-4 py-2 text-left font-medium text-gray-700">Parte</th>
                  <th className="border border-gray-200 px-4 py-2 text-left font-medium text-gray-700">Papel</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td className="border border-gray-200 px-4 py-2 font-medium">Instituto Itinerante de Tecnologia (IIT)</td>
                  <td className="border border-gray-200 px-4 py-2 text-gray-600">Operadora do marketplace — intermediária que fornece infraestrutura tecnológica</td>
                </tr>
                <tr className="bg-gray-50">
                  <td className="border border-gray-200 px-4 py-2 font-medium">Vendedor</td>
                  <td className="border border-gray-200 px-4 py-2 text-gray-600">Pessoa física ou jurídica que cria uma loja e oferta produtos/serviços</td>
                </tr>
                <tr>
                  <td className="border border-gray-200 px-4 py-2 font-medium">Comprador</td>
                  <td className="border border-gray-200 px-4 py-2 text-gray-600">Pessoa que realiza a compra de produtos listados por Vendedores</td>
                </tr>
              </tbody>
            </table>
          </div>
          <p className="text-gray-600 leading-relaxed mt-3">
            <strong>A IIT não é produtora nem vendedora dos produtos</strong> — é intermediária conforme CDC Art. 3º §2º.
            A responsabilidade principal pelos produtos é do Vendedor.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-xl font-semibold text-gray-800 mb-3">3. Responsabilidade do Vendedor</h2>
          <p className="text-gray-600 leading-relaxed mb-2">O Vendedor é responsável por:</p>
          <ul className="list-disc list-inside space-y-1 text-gray-600">
            <li>Qualidade e conformidade dos produtos com as descrições anunciadas (CDC Art. 18)</li>
            <li>Entrega no prazo e condições acordadas</li>
            <li>Informações verídicas sobre preço, disponibilidade e características</li>
            <li>Emissão de documentação fiscal de cada venda</li>
            <li>Conformidade legal do produto (proibido vender produtos ilegais ou falsificados)</li>
            <li>Atendimento pós-venda, incluindo trocas e devoluções conforme CDC</li>
          </ul>
        </section>

        <section className="mb-8">
          <h2 className="text-xl font-semibold text-gray-800 mb-3">4. Responsabilidade da Plataforma</h2>
          <p className="text-gray-600 leading-relaxed">
            A IIT pode ser responsabilizada subsidiariamente (CDC Art. 14) quando notificada judicialmente e não agir.
            A IIT responde diretamente por falhas na plataforma, segurança dos dados pessoais e indisponibilidade do serviço de pagamento.
            A IIT <strong>não responde</strong> por qualidade dos produtos de terceiros, atrasos de entrega ou extravio.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-xl font-semibold text-gray-800 mb-3">5. Direito de Arrependimento (CDC Art. 49)</h2>
          <p className="text-gray-600 leading-relaxed">
            O Comprador tem <strong>7 dias corridos</strong> a partir do recebimento para desistir da compra, sem necessidade de justificativa.
            O reembolso é integral, incluindo frete, em até <strong>15 dias</strong> após a comunicação. Para exercer esse direito,
            contate o Vendedor ou abra uma disputa pela plataforma.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-xl font-semibold text-gray-800 mb-3">6. Garantia Legal (CDC Art. 26)</h2>
          <ul className="list-disc list-inside space-y-1 text-gray-600">
            <li>Produtos não duráveis: 30 dias de garantia contra vícios aparentes</li>
            <li>Produtos duráveis: 90 dias de garantia contra vícios aparentes</li>
            <li>Vícios ocultos: prazo conta a partir da descoberta</li>
          </ul>
        </section>

        <section className="mb-8">
          <h2 className="text-xl font-semibold text-gray-800 mb-3">7. Produtos Proibidos</h2>
          <p className="text-gray-600 leading-relaxed mb-2">É vedado listar:</p>
          <ul className="list-disc list-inside space-y-1 text-gray-600">
            <li>Produtos falsificados, contrabandeados ou sem documentação fiscal</li>
            <li>Medicamentos sujeitos a controle sem prescrição</li>
            <li>Armas, munições ou itens de uso restrito</li>
            <li>Conteúdo que viole direitos autorais ou de imagem</li>
            <li>Qualquer produto proibido pela legislação brasileira</li>
          </ul>
          <p className="text-gray-600 leading-relaxed mt-2">
            A IIT pode remover produtos ou suspender lojas a qualquer momento por violação desta política.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-xl font-semibold text-gray-800 mb-3">8. Processo de Disputa</h2>
          <p className="text-gray-600 leading-relaxed">
            Em caso de conflito, as partes têm 5 dias úteis para resolução direta. Se não houver acordo,
            qualquer parte pode abrir uma Disputa Oficial na plataforma. A IIT analisará em até 10 dias úteis
            e poderá mediar, determinar reembolso ou suspender o Vendedor. Isso não impede o acesso às vias judiciais.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-xl font-semibold text-gray-800 mb-3">9. Taxas e Pagamentos</h2>
          <p className="text-gray-600 leading-relaxed">
            O StoreMake cobra uma taxa de plataforma sobre cada transação, conforme o plano do Vendedor.
            Repasses são feitos em até 2 dias úteis após confirmação do pagamento, descontada a taxa.
            Em caso de estorno ou chargeback, o Vendedor arca com o valor do produto e taxas bancárias.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-xl font-semibold text-gray-800 mb-3">10. Privacidade e LGPD</h2>
          <p className="text-gray-600 leading-relaxed">
            O tratamento de dados pessoais segue nossa Política de Privacidade.
            Dados dos Compradores são compartilhados com o Vendedor apenas para fins de entrega.
            O Vendedor assume a condição de Controlador de Dados em relação aos seus clientes e deve cumprir a LGPD.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-xl font-semibold text-gray-800 mb-3">11. Lei Aplicável e Foro</h2>
          <p className="text-gray-600 leading-relaxed">
            Estes Termos são regidos pela legislação brasileira (CDC, LGPD, Marco Civil da Internet).
            Fica eleito o Foro da Comarca de <strong>Salvador/BA</strong> para resolução de conflitos,
            salvo disposição legal em contrário.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-gray-800 mb-3">12. Contato</h2>
          <ul className="text-gray-600 space-y-1">
            <li><strong>Email:</strong> legal@institutoitinerante.com.br</li>
            <li><strong>DPO (LGPD):</strong> dpo@institutoitinerante.com.br</li>
            <li><strong>Salvador/BA, Brasil</strong></li>
          </ul>
        </section>
      </div>
    </div>
  )
}
